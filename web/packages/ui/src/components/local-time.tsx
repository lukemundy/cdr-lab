// No timeZone: Intl uses the time zone of the browser.
const dateTimeFormat = (seconds: boolean) =>
  new Intl.DateTimeFormat("en-AU", {
    day: "numeric",
    month: "long",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
    second: seconds ? "2-digit" : undefined,
    hour12: true,
  });

const formats = { minutes: dateTimeFormat(false), seconds: dateTimeFormat(true) };

// The parts make one shape in each browser: the literals between them differ in each ICU version.
function format(date: Date, seconds: boolean): string {
  const parts = Object.fromEntries(
    (seconds ? formats.seconds : formats.minutes)
      .formatToParts(date)
      .map(({ type, value }) => [type, value]),
  );
  const time = [parts.hour, parts.minute, ...(seconds ? [parts.second] : [])].join(":");
  return `${parts.day} ${parts.month} ${parts.year}, ${time} ${parts.dayPeriod?.toLowerCase()}`;
}

const plural = (count: number, unit: string) => `${count} ${unit}${count === 1 ? "" : "s"}`;

// The offset of the browser time zone at this instant, for example "UTC+8" and "UTC plus 8 hours".
function offset(date: Date): { visible: string; spoken: string } | undefined {
  const minutesEast = -date.getTimezoneOffset();
  if (minutesEast === 0) return undefined;
  const hours = Math.floor(Math.abs(minutesEast) / 60);
  const minutes = Math.abs(minutesEast) % 60;
  const visibleMinutes = minutes ? `:${String(minutes).padStart(2, "0")}` : "";
  const spokenMinutes = minutes ? ` ${plural(minutes, "minute")}` : "";
  return {
    visible: `UTC${minutesEast > 0 ? "+" : "-"}${hours}${visibleMinutes}`,
    spoken: `UTC ${minutesEast > 0 ? "plus" : "minus"} ${plural(hours, "hour")}${spokenMinutes}`,
  };
}

export type LocalTimeProps = {
  // An ISO 8601 time.
  value: string;
  seconds?: boolean;
};

// A screen reader can read "UTC-3" as "UTC 3", so it gets the offset in words.
export function LocalTime({ value, seconds = false }: LocalTimeProps) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return <>{value}</>;
  const zone = offset(date);
  return (
    <time dateTime={date.toISOString()}>
      {format(date, seconds)}{" "}
      {zone ? (
        <>
          <span aria-hidden="true">{zone.visible}</span>
          <span className="sr-only select-none">{zone.spoken}</span>
        </>
      ) : (
        "UTC"
      )}
    </time>
  );
}
