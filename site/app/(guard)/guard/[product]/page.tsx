import { ProductDocs } from '@/components/guard/ProductDocs';

/**
 * The bare product URL renders the Documentation tab.
 *
 * Not a redirect to /docs: a redirect costs a round trip and puts a URL in the
 * reader's history they did not ask for. Both paths render the same component,
 * and ProductTabs marks Documentation active for either one.
 */
export default async function ProductPage({
  params,
}: {
  params: Promise<{ product: string }>;
}) {
  const { product } = await params;
  return <ProductDocs slug={product} />;
}
