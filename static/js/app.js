const rupiah = value => new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(value);

document.addEventListener("click", event => {
  const remove = event.target.closest(".row-remove");
  if (remove && remove.closest("[data-list]").children.length > 1) remove.closest(".repeat-row").remove();
  const add = event.target.closest("[data-add-row]");
  if (add) {
    const list = document.querySelector(`[data-list="${add.dataset.addRow}"]`);
    const row = list.firstElementChild.cloneNode(true);
    row.querySelectorAll("input").forEach(input => input.value = input.type === "number" ? "1" : "");
    list.append(row);
  }
  const product = event.target.closest("[data-add-cart]");
  if (product) addCart(...product.dataset.addCart.split("|"));
  if (event.target.closest("[data-clear-cart]")) { cart.length = 0; renderCart(); }
  if (event.target.closest("[data-print]")) window.print();
});

const cart = [];
function addCart(name, unit, price) {
  const found = cart.find(item => item.name === name && item.unit === unit);
  found ? found.qty++ : cart.push({ name, unit, price: Number(price), qty: 1 });
  renderCart();
}
function renderCart() {
  const list = document.querySelector("[data-cart-list]");
  if (!list) return;
  list.innerHTML = cart.length ? cart.map((item, i) => `<div class="cart-item"><div><strong>${item.name}</strong><small>${item.unit} · ${rupiah(item.price)}</small></div><div class="qty"><button type="button" data-cart-minus="${i}">−</button><span>${item.qty}</span><button type="button" data-cart-plus="${i}">+</button></div><b>${rupiah(item.qty * item.price)}</b></div>`).join("") : '<div class="cart-empty">Pilih obat untuk menambahkannya.</div>';
  const total = cart.reduce((sum, item) => sum + item.qty * item.price, 0);
  document.querySelector("[data-cart-total]").textContent = rupiah(total);
  updateChange(total);
}
document.addEventListener("click", event => {
  const plus = event.target.closest("[data-cart-plus]");
  const minus = event.target.closest("[data-cart-minus]");
  if (plus) cart[Number(plus.dataset.cartPlus)].qty++;
  if (minus) {
    const index = Number(minus.dataset.cartMinus);
    cart[index].qty--;
    if (!cart[index].qty) cart.splice(index, 1);
  }
  if (plus || minus) renderCart();
});
document.querySelector("[data-payment]")?.addEventListener("input", () => updateChange(cart.reduce((sum, item) => sum + item.qty * item.price, 0)));
function updateChange(total) {
  const payment = Number(document.querySelector("[data-payment]")?.value || 0);
  const change = document.querySelector("[data-change]");
  if (change) change.textContent = rupiah(Math.max(0, payment - total));
}

document.addEventListener("input", event => {
  if (!event.target.matches(".opname-count")) return;
  const diff = Number(event.target.value) - Number(event.target.dataset.system);
  const cell = event.target.closest("tr").querySelector(".stock-diff");
  cell.textContent = diff > 0 ? `+${diff}` : String(diff).replace("-", "−");
  cell.className = `stock-diff ${diff > 0 ? "text-success" : diff < 0 ? "text-danger" : ""}`;
  const reason = event.target.closest("tr").querySelector("td:last-child input");
  reason.disabled = diff === 0;
  if (!diff) reason.value = "";
});
