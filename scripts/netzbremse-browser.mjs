import puppeteer from "puppeteer";

const url = process.env.NB_SPEEDTEST_URL || "https://netzbremse.de/speed";
const acceptedPrivacyPolicy =
  process.env.NB_SPEEDTEST_ACCEPT_POLICY?.toLowerCase() === "true";

if (!acceptedPrivacyPolicy) {
  console.error(
    'NB_SPEEDTEST_ACCEPT_POLICY="true" is required before running the netzbremse browser flow.',
  );
  process.exit(1);
}

const browser = await puppeteer.launch({
  ...(process.env.PUPPETEER_EXECUTABLE_PATH
    ? { executablePath: process.env.PUPPETEER_EXECUTABLE_PATH }
    : {}),
  headless: true,
  pipe: true,
  userDataDir: process.env.NB_SPEEDTEST_BROWSER_DATA_DIR || "/tmp/netzbremse-browser",
  args: [
    "--no-sandbox",
    "--disable-setuid-sandbox",
    "--disable-dev-shm-usage",
    "--disable-gpu",
    "--no-zygote",
  ],
});

try {
  const page = await browser.newPage();
  await page.setViewport({ width: 1100, height: 1200 });
  await page.goto(url, { waitUntil: "domcontentloaded" });
  await page.waitForSelector("nb-speedtest >>>> #nb_speedtest_start_btn", {
    timeout: 60000,
  });

  await page.evaluate(() => {
    window.nbSpeedtestOptions = { acceptedPolicy: true };
  });

  const resultPromise = new Promise((resolve) => {
    page.exposeFunction("nbSpeedtestOnResult", (payload) => resolve(payload));
  });
  const finishedPromise = new Promise((resolve) => {
    page.exposeFunction("nbSpeedtestOnFinished", () => resolve());
  });
  await page.click("nb-speedtest >>>> #nb_speedtest_start_btn");

  const result = await resultPromise;
  await finishedPromise;
  process.stdout.write(JSON.stringify(result));
} finally {
  await browser.close();
}
