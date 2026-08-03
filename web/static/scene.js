/* Three.js background for the portfolio.
 *
 * A real WebGL scene rather than a CSS gradient: a rotating particle sphere
 * with a wireframe core, parallaxed by the pointer and rotated by scroll.
 *
 * Design constraints that shaped this:
 *  - It must never block content. The canvas is fixed, z-index 0,
 *    pointer-events:none, and every failure path degrades to a plain
 *    dark background rather than a broken page.
 *  - It must not cook a laptop battery. The renderer pauses when the tab is
 *    hidden, and pixel ratio is capped at 2 - beyond that the extra pixels
 *    are invisible but the fill cost is real.
 *  - It must respect prefers-reduced-motion by not running at all.
 */
(function () {
  'use strict';

  var host = document.getElementById('scene');
  if (!host) return;

  // Honour the OS accessibility setting before spending anything on WebGL.
  if (window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;

  // Some machines and locked-down browsers have no WebGL. Bail silently -
  // the CSS background already looks intentional on its own.
  if (typeof THREE === 'undefined') return;

  var accent = new THREE.Color(host.dataset.accent || '#4f8cff');
  var accent2 = new THREE.Color(host.dataset.accent2 || '#22d3ee');

  var renderer;
  try {
    renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
  } catch (e) {
    return;
  }
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
  renderer.setSize(window.innerWidth, window.innerHeight);
  host.appendChild(renderer.domElement);

  var scene = new THREE.Scene();
  scene.fog = new THREE.FogExp2(0x04060c, 0.055);

  var camera = new THREE.PerspectiveCamera(58, window.innerWidth / window.innerHeight, 0.1, 100);
  camera.position.z = 15;

  var world = new THREE.Group();
  scene.add(world);

  // ── Particle shell ────────────────────────────────────────────────
  // Points distributed on a sphere using the golden-angle spiral. Naive
  // random spherical coordinates bunch up at the poles and look obviously
  // wrong; this gives an even shell.
  var COUNT = 2600;
  var positions = new Float32Array(COUNT * 3);
  var colors = new Float32Array(COUNT * 3);
  var golden = Math.PI * (3 - Math.sqrt(5));

  for (var i = 0; i < COUNT; i++) {
    var y = 1 - (i / (COUNT - 1)) * 2;
    var radiusAtY = Math.sqrt(Math.max(0, 1 - y * y));
    var theta = golden * i;

    // Slight radial jitter so it reads as a cloud rather than a hard shell.
    var r = 6.2 + (Math.random() - 0.5) * 1.5;

    positions[i * 3] = Math.cos(theta) * radiusAtY * r;
    positions[i * 3 + 1] = y * r;
    positions[i * 3 + 2] = Math.sin(theta) * radiusAtY * r;

    var mixed = accent.clone().lerp(accent2, (y + 1) / 2);
    colors[i * 3] = mixed.r;
    colors[i * 3 + 1] = mixed.g;
    colors[i * 3 + 2] = mixed.b;
  }

  var pGeo = new THREE.BufferGeometry();
  pGeo.setAttribute('position', new THREE.BufferAttribute(positions, 3));
  pGeo.setAttribute('color', new THREE.BufferAttribute(colors, 3));

  var points = new THREE.Points(pGeo, new THREE.PointsMaterial({
    size: 0.075,
    vertexColors: true,
    transparent: true,
    opacity: 0.9,
    depthWrite: false,
    blending: THREE.AdditiveBlending
  }));
  world.add(points);

  // ── Wireframe core ────────────────────────────────────────────────
  var core = new THREE.Mesh(
    new THREE.IcosahedronGeometry(3.1, 1),
    new THREE.MeshBasicMaterial({
      color: accent, wireframe: true, transparent: true, opacity: 0.24
    })
  );
  world.add(core);

  var coreInner = new THREE.Mesh(
    new THREE.IcosahedronGeometry(1.7, 0),
    new THREE.MeshBasicMaterial({
      color: accent2, wireframe: true, transparent: true, opacity: 0.34
    })
  );
  world.add(coreInner);

  // ── Drifting satellites ───────────────────────────────────────────
  var sats = [];
  for (var s = 0; s < 5; s++) {
    var m = new THREE.Mesh(
      new THREE.TetrahedronGeometry(0.42 + Math.random() * 0.3),
      new THREE.MeshBasicMaterial({
        color: s % 2 ? accent2 : accent, wireframe: true, transparent: true, opacity: 0.5
      })
    );
    m.userData = {
      radius: 8 + Math.random() * 3,
      speed: 0.12 + Math.random() * 0.22,
      phase: Math.random() * Math.PI * 2,
      tilt: (Math.random() - 0.5) * 1.4
    };
    world.add(m);
    sats.push(m);
  }

  // ── Interaction ───────────────────────────────────────────────────
  var pointer = { x: 0, y: 0 };
  var target = { x: 0, y: 0 };
  var scrollNorm = 0;

  window.addEventListener('pointermove', function (e) {
    // Normalised to -0.5..0.5 so the parallax feels the same on any screen size.
    target.x = e.clientX / window.innerWidth - 0.5;
    target.y = e.clientY / window.innerHeight - 0.5;
  }, { passive: true });

  window.addEventListener('scroll', function () {
    var max = document.body.scrollHeight - window.innerHeight;
    scrollNorm = max > 0 ? window.scrollY / max : 0;
  }, { passive: true });

  window.addEventListener('resize', function () {
    camera.aspect = window.innerWidth / window.innerHeight;
    camera.updateProjectionMatrix();
    renderer.setSize(window.innerWidth, window.innerHeight);
  });

  // A background animation on a hidden tab is pure waste - it keeps the GPU
  // awake and drains battery for something nobody is looking at.
  var running = true;
  document.addEventListener('visibilitychange', function () {
    running = !document.hidden;
    if (running) loop();
  });

  var clock = new THREE.Clock();

  function loop() {
    if (!running) return;
    requestAnimationFrame(loop);

    var t = clock.getElapsedTime();

    // Ease toward the pointer instead of snapping to it. The lag is what
    // makes the movement feel like weight rather than a jitter.
    pointer.x += (target.x - pointer.x) * 0.045;
    pointer.y += (target.y - pointer.y) * 0.045;

    world.rotation.y = t * 0.05 + pointer.x * 0.55 + scrollNorm * Math.PI * 0.8;
    world.rotation.x = pointer.y * 0.38 + scrollNorm * 0.35;

    core.rotation.y = -t * 0.13;
    core.rotation.x = t * 0.08;
    coreInner.rotation.y = t * 0.26;
    coreInner.rotation.z = -t * 0.17;

    for (var i = 0; i < sats.length; i++) {
      var d = sats[i].userData;
      var a = t * d.speed + d.phase;
      sats[i].position.set(
        Math.cos(a) * d.radius,
        Math.sin(a * 0.7) * d.radius * 0.35 + d.tilt * 2,
        Math.sin(a) * d.radius
      );
      sats[i].rotation.x = a * 1.5;
      sats[i].rotation.y = a * 1.1;
    }

    // Gentle breathing so the scene is never completely static.
    camera.position.z = 15 + Math.sin(t * 0.25) * 0.7;

    renderer.render(scene, camera);
  }

  loop();
})();

/* Scroll reveal, shared by every page. */
(function () {
  var items = document.querySelectorAll('.rv');
  if (!('IntersectionObserver' in window)) {
    // Without the observer, show everything. A blank page is a far worse
    // failure than a page with no animation.
    for (var i = 0; i < items.length; i++) items[i].classList.add('in');
    return;
  }
  var io = new IntersectionObserver(function (entries) {
    entries.forEach(function (e) {
      if (e.isIntersecting) { e.target.classList.add('in'); io.unobserve(e.target); }
    });
  }, { threshold: 0.1 });
  items.forEach(function (el) { io.observe(el); });
})();

/* Card tilt. Rotates a card toward the cursor in real 3D. */
(function () {
  if (window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
  // Touch devices have no hover, and a tilt that only fires on tap feels broken.
  if (!window.matchMedia || !window.matchMedia('(hover: hover)').matches) return;

  document.querySelectorAll('.tilt').forEach(function (card) {
    card.addEventListener('pointermove', function (e) {
      var r = card.getBoundingClientRect();
      var px = (e.clientX - r.left) / r.width - 0.5;
      var py = (e.clientY - r.top) / r.height - 0.5;
      // Capped at 9 degrees. Beyond that the text starts to distort and the
      // effect reads as a gimmick instead of depth.
      card.style.transform =
        'perspective(900px) rotateY(' + (px * 9) + 'deg) rotateX(' + (-py * 9) + 'deg) translateY(-5px)';
    });
    card.addEventListener('pointerleave', function () {
      card.style.transform = '';
    });
  });
})();
