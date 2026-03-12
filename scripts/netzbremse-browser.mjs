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
  headless: true,
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
  await page.waitForNetworkIdle();

  await page.evaluate(() => {
    window.nbSpeedtestOptions = { acceptedPolicy: true };
  });

  const result = await new Promise(async (resolve, reject) => {
    await page.exposeFunction("nbSpeedtestOnResult", (payload) => resolve(payload));
    await page.exposeFunction("nbSpeedtestOnFinished", () => {});

    try {
      await page.click("nb-speedtest >>>> #nb_speedtest_start_btn");
    } catch (error) {
      reject(error);
    }
  });

  process.stdout.write(JSON.stringify(result));
} finally {
  await browser.close();
}
