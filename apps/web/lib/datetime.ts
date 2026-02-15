export type DateTimeInput = Date | string | number | null | undefined;

export function formatLocaleDateTime(
  value: DateTimeInput,
  locale?: string,
  options: Intl.DateTimeFormatOptions = { dateStyle: "medium", timeStyle: "medium" }
): string {
  if (value === null || value === undefined) {
    return "-";
  }

  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }

  const safeLocale = locale?.trim() || undefined;
  const formatter = new Intl.DateTimeFormat(safeLocale, options);
  return formatter.format(date);
}

export function formatLocaleTime(
  value: DateTimeInput,
  locale?: string,
  options: Intl.DateTimeFormatOptions = { timeStyle: "medium" }
): string {
  return formatLocaleDateTime(value, locale, options);
}

