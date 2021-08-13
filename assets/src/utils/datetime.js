import dayjs from "dayjs";
import timezone from "dayjs/plugin/timezone";
import utc from "dayjs/plugin/utc";
import Auth from "../middleware/Auth";
dayjs.extend(utc);
dayjs.extend(timezone);

const defaultTimeZone = "Asia/Shanghai";
const preferTimeZone = Auth.GetPreference("timeZone");
export let timeZone = preferTimeZone ? preferTimeZone : defaultTimeZone;

export function refreshTimeZone() {
    timeZone = Auth.GetPreference("timeZone");
    timeZone = timeZone ? timeZone : defaultTimeZone;
}

export function formatLocalTime(time, format) {
    return dayjs(time).tz(timeZone).format(format);
}

export function validateTimeZone(name) {
    try {
        dayjs().tz(name).format();
    } catch (e) {
        return false;
    }
    return true;
}
