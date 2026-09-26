const $ = selector => document.querySelector(selector);
const worker = new Worker("worker.js");
const pending = new Map();
let requestID = 0;
let heroVersion = 0;
let batchVersion = 0;
let prefix = "robot";

worker.onmessage = ({data: {id, data, error}}) => {
  const request = pending.get(id);
  if (!request) return;
  pending.delete(id);
  error ? request.reject(new Error(error)) : request.resolve(data);
};
worker.onerror = event => {
  for (const request of pending.values()) request.reject(new Error(event.message || "The renderer stopped. Reload to try again."));
  pending.clear();
};
function render(seed, size, format = "svg", transparent = $("#transparent").checked) {
  return new Promise((resolve, reject) => {
    const id = ++requestID;
    pending.set(id, {resolve, reject});
    worker.postMessage({id, seed, size, format, transparent});
  });
}
function avatar(svg, size, label = "") {
  const box = document.createElement("div");
  box.className = "avatar";
  box.style.width = box.style.height = `${size}px`;
  const image = document.createElement("img");
  image.src = `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`;
  image.alt = label;
  box.append(image);
  return box;
}
function report(error) { $("#status").textContent = `Something went wrong: ${error.message}`; }

async function updateHero() {
  const version = ++heroVersion;
  const seed = $("#seed").value;
  const svg = await render(seed, 256);
  if (version !== heroVersion) return;
  const image = avatar(svg, 224, `Robot ${seed}`).firstChild;
  $("#hero").replaceChildren(image);
  $("#hero-seed").textContent = seed || "(empty seed)";
  $("#sizes").replaceChildren(...[24, 32, 64].map(size => {
    const figure = document.createElement("figure");
    const caption = document.createElement("figcaption");
    caption.textContent = `${size} px`;
    figure.append(avatar(svg, size), caption);
    return figure;
  }));
}

async function updateGallery() {
  const version = ++batchVersion;
  const batchPrefix = prefix;
  const transparent = $("#transparent").checked;
  $("#gallery").replaceChildren();
  for (let i = 0; i < 200; i++) {
    if (version !== batchVersion) return;
    const seed = `${batchPrefix}-${String(i).padStart(3, "0")}`;
    const svg = await render(seed, 128, "svg", transparent);
    if (version !== batchVersion) return;
    const card = document.createElement("button");
    card.className = "card";
    card.type = "button";
    card.setAttribute("aria-label", `Choose ${seed}`);
    const label = document.createElement("code");
    label.textContent = seed;
    card.append(avatar(svg, 112), label);
    card.addEventListener("click", () => {
      $("#seed").value = seed;
      updateHero().then(() => $(".workbench").scrollIntoView({behavior: "auto", block: "start"})).catch(report);
    });
    $("#gallery").append(card);
    $("#status").textContent = `${i + 1} / 200 robots ready`;
  }
  $("#status").textContent = "200 robots · generated live in your browser";
}

async function download(format) {
  const button = $(`#${format}`);
  button.disabled = true;
  try {
    const seed = $("#seed").value;
    const size = Number($("#size").value);
    const data = await render(seed, size, format);
    const content = format === "png" ? Uint8Array.from(atob(data), c => c.charCodeAt(0)) : data;
    const blob = new Blob([content], {type: format === "png" ? "image/png" : "image/svg+xml"});
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `roboticon-${size}.${format}`;
    link.click();
    setTimeout(() => URL.revokeObjectURL(url), 10000);
  } finally { button.disabled = false; }
}

let debounce;
$("#seed").addEventListener("input", () => {
  ++heroVersion;
  clearTimeout(debounce);
  debounce = setTimeout(() => updateHero().catch(report), 120);
});
$("#transparent").addEventListener("change", () => {
  updateHero().catch(report);
  updateGallery().catch(report);
});
$("#circle").addEventListener("change", event => document.body.classList.toggle("circle-preview", event.target.checked));
$("#shuffle").addEventListener("click", () => {
  $("#seed").value = `friend-${crypto.randomUUID().slice(0, 8)}`;
  updateHero().catch(report);
});
$("#batch").addEventListener("click", () => {
  prefix = `crew-${crypto.randomUUID().slice(0, 5)}`;
  updateGallery().catch(report);
});
for (const format of ["png", "svg"]) $(`#${format}`).addEventListener("click", () => download(format).catch(report));

try {
  await updateHero();
  for (const id of ["shuffle", "batch", "png", "svg"]) $(`#${id}`).disabled = false;
  await updateGallery();
} catch (error) { report(error); }
