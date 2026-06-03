import { readFile } from "node:fs/promises";
import assert from "node:assert/strict";

const html = await readFile(new URL("../index.html", import.meta.url), "utf8");
const js = await readFile(new URL("../src/app.js", import.meta.url), "utf8");
const css = await readFile(new URL("../src/styles.css", import.meta.url), "utf8");

assert.match(html, /End-to-end encrypted/i, "chat must be positioned as E2EE only");
assert.match(html, /Голосовое/, "voice-message control must be present");
assert.match(html, /Экспорт HTML/, "HTML export control must be present");
assert.match(html, /Код с почты/, "email-code authentication must be present");
assert.match(html, /пароль/, "password authentication must be present");

assert.match(js, /const RETENTION_DAYS = 180/, "message retention must be at least 180 days");
assert.match(js, /AES-GCM/, "messages must use Web Crypto encryption before storage");
assert.match(js, /MediaRecorder/, "voice messages must use MediaRecorder");
assert.doesNotMatch(html, /type="file"/, "photo and video uploads must not be available");

assert.match(css, /--bg: #06070b/, "Raven design must use a dark palette");
assert.match(css, /--accent: #8b5cf6/, "Raven design must include violet accent color");

console.log("Raven Chat static contract checks passed.");
