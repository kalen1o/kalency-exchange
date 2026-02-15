export function formatFloatPrice(
  value: number | null | undefined,
  currency: string,
  locale = "en-US"
): string {
  if (value === null || value === undefined || Number.isNaN(value)) {
    return "-";
  }

  const safeCurrency = currency.trim().toUpperCase() || "USD";
  const formatter = new Intl.NumberFormat(locale, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  });
  return `${formatter.format(value)} ${safeCurrency}`;
}
