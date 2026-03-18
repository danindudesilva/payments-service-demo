(() => {
  const config = window.__DEMO_CONFIG__ || {};
  const stripePublishableKey = config.stripePublishableKey || "";

  const output = document.getElementById("output");
  const createForm = document.getElementById("create-form");
  const paymentPanel = document.getElementById("payment-panel");
  const confirmButton = document.getElementById("confirm-button");
  const fetchButton = document.getElementById("fetch-button");
  const reconcileButton = document.getElementById("reconcile-button");
  const resetButton = document.getElementById("reset-button");

  const orderIdInput = document.getElementById("order-id");
  const amountInput = document.getElementById("amount");
  const currencyInput = document.getElementById("currency");

  const stateAttemptId = document.getElementById("state-attempt-id");
  const stateProviderPaymentId = document.getElementById(
    "state-provider-payment-id"
  );
  const stateStatus = document.getElementById("state-status");
  const stateFailureReason = document.getElementById("state-failure-reason");
  const stateClientSecret = document.getElementById("state-client-secret");
  const stateReturnURL = document.getElementById("state-return-url");
  const stateReturned = document.getElementById("state-returned");

  const stripe = stripePublishableKey ? Stripe(stripePublishableKey) : null;

  let elements = null;
  let paymentElement = null;
  let attemptId = null;
  let clientSecret = null;
  let lastPayload = null;

  const STORAGE_KEY = "payments_service_demo_state";

  function log(value) {
    output.textContent =
      typeof value === "string" ? value : JSON.stringify(value, null, 2);
  }

  function generateDemoSuffix() {
    return Math.random().toString(36).slice(2, 10);
  }

  function generateDefaultOrderId() {
    return `order_demo_${generateDemoSuffix()}`;
  }

  function getAttemptIdFromURL() {
    const params = new URLSearchParams(window.location.search);
    return params.get("attempt_id");
  }

  function setAttemptIdInURL(id) {
    const url = new URL(window.location.href);
    url.searchParams.set("attempt_id", id);
    window.history.replaceState({}, "", url.toString());
  }

  function clearAttemptIdFromURL() {
    const url = new URL(window.location.href);
    url.searchParams.delete("attempt_id");
    window.history.replaceState({}, "", url.toString());
  }

  function getReturnURL(id) {
    const url = new URL(`${window.location.origin}/demo`);
    url.searchParams.set("attempt_id", id);
    return url.toString();
  }

  function returnedFromStripe() {
    return Boolean(getAttemptIdFromURL());
  }

  function persistState() {
    const state = {
      attemptId,
      clientSecret,
      lastPayload,
    };

    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  }

  function restoreState() {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return;
    }

    try {
      const state = JSON.parse(raw);
      attemptId = state.attemptId || attemptId;
      clientSecret = state.clientSecret || clientSecret;
      lastPayload = state.lastPayload || lastPayload;
    } catch (error) {
      console.warn("failed to restore demo state", error);
    }
  }

  function clearState() {
    attemptId = null;
    clientSecret = null;
    lastPayload = null;
    orderIdInput.value = generateDefaultOrderId();
    if (paymentElement) {
      paymentElement.destroy();
      paymentElement = null;
    }

    elements = null;
    paymentPanel.classList.add("hidden");
    sessionStorage.removeItem(STORAGE_KEY);
    clearAttemptIdFromURL();
    renderState();
    log("Demo state cleared.");
  }

  function renderState() {
    stateAttemptId.textContent = attemptId || "-";
    stateProviderPaymentId.textContent =
      lastPayload?.provider?.payment_id || "-";
    stateStatus.textContent = lastPayload?.status || "-";
    stateFailureReason.textContent = lastPayload?.failure_reason || "-";
    stateClientSecret.textContent = clientSecret ? "present" : "not loaded";
    stateReturnURL.textContent = attemptId ? getReturnURL(attemptId) : "-";
    stateReturned.textContent = returnedFromStripe() ? "yes" : "no";
    orderIdInput.value = generateDefaultOrderId();
  }

  async function createAttempt({ orderId, amount, currency }) {
    const response = await fetch("/payment-attempts", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Idempotency-Key": `demo-${orderId}`,
      },
      body: JSON.stringify({
        order_id: orderId,
        amount,
        currency,
        return_url: getReturnURL(orderId), // kept for API contract; see note below
        description: "payments-service demo payment",
      }),
    });

    const body = await response.json();

    if (!response.ok) {
      throw new Error(body.error || "failed to create payment attempt");
    }

    return body;
  }

  async function fetchAttempt(id) {
    const response = await fetch(
      `/payment-attempts/${encodeURIComponent(id)}`,
      {
        method: "GET",
      }
    );

    const body = await response.json();

    if (!response.ok) {
      throw new Error(body.error || "failed to fetch payment attempt");
    }

    return body;
  }

  async function reconcileAttempt(id) {
    const response = await fetch(
      `/payment-attempts/${encodeURIComponent(id)}/reconcile`,
      {
        method: "POST",
      }
    );

    const body = await response.json();

    if (!response.ok) {
      throw new Error(body.error || "failed to reconcile payment attempt");
    }

    return body;
  }

  function mountPaymentElement(secret) {
    if (!stripe) {
      throw new Error("Stripe publishable key is not configured.");
    }

    if (paymentElement) {
      paymentElement.destroy();
      paymentElement = null;
    }

    elements = stripe.elements({ clientSecret: secret });
    paymentElement = elements.create("payment");
    paymentElement.mount("#payment-element");
    paymentPanel.classList.remove("hidden");
  }

  async function handleCreateSubmit(event) {
    event.preventDefault();

    try {
      const orderId = orderIdInput.value.trim();
      const amount = parseMajorAmountToMinorUnits(amountInput.value);
      const currency = currencyInput.value.trim();

      const attempt = await createAttempt({ orderId, amount, currency });

      attemptId = attempt.id;
      clientSecret = attempt.provider.client_secret;
      lastPayload = attempt;

      setAttemptIdInURL(attemptId);
      persistState();
      renderState();
      mountPaymentElement(clientSecret);
      log(attempt);
    } catch (error) {
      log(error.message || String(error));
    }
  }

  async function handleConfirmClick() {
    try {
      if (!stripe || !elements || !clientSecret || !attemptId) {
        log("Create a payment attempt first.");
        return;
      }

      const result = await stripe.confirmPayment({
        elements,
        confirmParams: {
          return_url: getReturnURL(attemptId),
        },
      });

      if (result.error) {
        log(result.error.message || "payment confirmation failed");
        return;
      }

      log("Payment confirmation submitted.");
    } catch (error) {
      log(error.message || String(error));
    }
  }

  async function handleFetchClick() {
    try {
      const id = attemptId || getAttemptIdFromURL();
      if (!id) {
        log("No attempt ID available yet.");
        return;
      }

      const attempt = await fetchAttempt(id);
      lastPayload = attempt;
      persistState();
      renderState();
      log(attempt);
    } catch (error) {
      log(error.message || String(error));
    }
  }

  async function handleReconcileClick() {
    try {
      const id = attemptId || getAttemptIdFromURL();
      if (!id) {
        log("No attempt ID available yet.");
        return;
      }

      const reconciled = await reconcileAttempt(id);
      lastPayload = reconciled;
      persistState();
      renderState();
      log(reconciled);
    } catch (error) {
      log(error.message || String(error));
    }
  }

  async function bootstrapFromURL() {
    restoreState();

    const idFromURL = getAttemptIdFromURL();
    if (idFromURL) {
      attemptId = idFromURL;
    }

    renderState();

    if (!attemptId) {
      return;
    }

    try {
      const reconciled = await reconcileAttempt(attemptId);
      lastPayload = reconciled;
      persistState();
      renderState();
      log(reconciled);
    } catch (error) {
      log(error.message || String(error));
    }
  }

  function parseMajorAmountToMinorUnits(value) {
    const normalized = String(value).trim();

    if (!normalized) {
      throw new Error("Amount is required.");
    }

    const parsed = Number(normalized);
    if (!Number.isFinite(parsed)) {
      throw new Error("Amount must be a valid number.");
    }

    if (parsed <= 0) {
      throw new Error("Amount must be greater than zero.");
    }

    const minorUnits = Math.round(parsed * 100);

    if (minorUnits <= 0) {
      throw new Error("Amount must be greater than zero.");
    }

    return minorUnits;
  }

  function formatMinorUnitsToMajor(minorUnits) {
    return (minorUnits / 100).toFixed(2);
  }

  createForm.addEventListener("submit", handleCreateSubmit);
  confirmButton.addEventListener("click", handleConfirmClick);
  fetchButton.addEventListener("click", handleFetchClick);
  reconcileButton.addEventListener("click", handleReconcileClick);
  resetButton.addEventListener("click", clearState);

  bootstrapFromURL();
})();
