// The rule of the consent page (services/dataholder/internal/adapter/consumerpages, durationText), so that the
// dashboard gives the period that the consumer saw at the consent.

const hour = 3600;
const hoursInDay = 24;
const daysInYear = 365;
// CDS: Requesting Sharing Duration. A sharing duration of 24 hours or less is a disclosure on one occasion.
const oneOccasionLimit = hoursInDay * hour;

function quantity(amount: number, unit: string): string {
  return `${amount} ${unit}${amount === 1 ? "" : "s"}`;
}

export function formatPeriod(sharingDurationSeconds: number): string {
  if (sharingDurationSeconds <= oneOccasionLimit) return "One occasion";
  const totalHours = Math.ceil(sharingDurationSeconds / hour);
  const days = Math.floor(totalHours / hoursInDay);
  const hours = totalHours % hoursInDay;
  if (days === daysInYear && hours === 0) return "1 year";
  const parts = [];
  if (days > 0) parts.push(quantity(days, "day"));
  if (hours > 0) parts.push(quantity(hours, "hour"));
  return parts.join(" and ");
}
