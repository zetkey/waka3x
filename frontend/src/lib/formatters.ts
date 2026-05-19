export function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }
  return `${minutes}m`;
}

type DateInput = Date | string | null | undefined;

const isoDateTimePattern = /^\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}/;
const timezoneSuffixPattern = /(?:z|[+-]\d{2}:?\d{2})$/i;

export function parseApiDate(date: DateInput): Date | null {
  if (!date) return null;
  if (date instanceof Date) return Number.isNaN(date.getTime()) ? null : date;

  const value = date.trim();
  if (!value) return null;

  const normalized =
    isoDateTimePattern.test(value) && !timezoneSuffixPattern.test(value)
      ? `${value.replace(" ", "T")}Z`
      : value;
  const parsed = new Date(normalized);

  return Number.isNaN(parsed.getTime()) ? null : parsed;
}

function localeOptions(
  options: Intl.DateTimeFormatOptions,
  timeZone?: string,
): Intl.DateTimeFormatOptions {
  if (!timeZone || timeZone === "Local") return options;

  try {
    new Intl.DateTimeFormat("en-US", { timeZone });
    return { ...options, timeZone };
  } catch {
    return options;
  }
}

export function formatDate(date: DateInput, timeZone?: string): string {
  return (
    parseApiDate(date)?.toLocaleDateString(
      "en-US",
      localeOptions(
        {
          month: "short",
          day: "numeric",
          year: "numeric",
        },
        timeZone,
      ),
    ) ?? "-"
  );
}

export function formatDateTime(date: DateInput, timeZone?: string): string {
  return (
    parseApiDate(date)?.toLocaleString(
      "en-US",
      localeOptions(
        {
          month: "short",
          day: "numeric",
          year: "numeric",
          hour: "2-digit",
          minute: "2-digit",
        },
        timeZone,
      ),
    ) ?? "-"
  );
}

export function formatTime(date: DateInput, timeZone?: string): string {
  return (
    parseApiDate(date)?.toLocaleTimeString(
      [],
      localeOptions(
        {
          hour: "2-digit",
          minute: "2-digit",
        },
        timeZone,
      ),
    ) ?? "-"
  );
}

export function formatWeekday(date: DateInput, timeZone?: string): string {
  return (
    parseApiDate(date)?.toLocaleDateString(
      "en-US",
      localeOptions(
        {
          weekday: "short",
        },
        timeZone,
      ),
    ) ?? "-"
  );
}
