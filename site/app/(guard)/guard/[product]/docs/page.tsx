import { ProductDocs } from '@/components/guard/ProductDocs';

/** Canonical documentation URL. Same content as the bare product page. */
export default async function ProductDocsPage({
  params,
}: {
  params: Promise<{ product: string }>;
}) {
  const { product } = await params;
  return <ProductDocs slug={product} />;
}
