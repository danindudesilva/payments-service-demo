(() => {
  const config = window.__DEMO_CONFIG__ || {};
  const stripePublishableKey = config.stripePublishableKey || "";

  const output = document.getElementById("output");
  const createForm = document.getElementById("create-form");
  const paymentPanel = document.getElementById("payment-panel");
  const confirmButton = document.getElementById("confirm-button");
  const reconcileButton = document.getElementById("reconcile-button");

  const orderIdInput = document.getElementById("order-id");
  const amountInput = document.getElementById("amount");
  const currencyInput = document.getElementById("currency");

  const stripe = stripePublishableKey ? Stripe(stripePublishableKey) : null;

  let elements = null;
  let paymentElement = null;
  let attemptId = null;
  let clientSecret = null;

  function log(value) {
    output.textContent =
      typeof value === "string" ? value : JSON.stringify(value, null, 2);
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

  function getReturnURL(id) {
    const url = new URL(`${window.location.origin}/demo`);
    url.searchParams.set("attempt_id", id);
    return url.toString();
  }

  async function createAttempt({ orderId, amount, currency }) {
    const response = await fetch("/payment-attempts", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        order_id: orderId,
        amount,
        currency,
        return_url: getReturnURL(orderId), // ideally
        description: "payments-service demo payment",
      }),
    });

    const body = await response.json();

    if (!response.ok) {
      throw new Error(body.error || "failed to create payment attempt");
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
      const amount = Number(amountInput.value);
      const currency = currencyInput.value.trim();

      const attempt = await createAttempt({ orderId, amount, currency });

      attemptId = attempt.id;
      clientSecret = attempt.provider.client_secret;

      setAttemptIdInURL(attemptId);
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

  async function handleReconcileClick() {
    try {
      const id = attemptId || getAttemptIdFromURL();

      if (!id) {
        log("No attempt ID available yet.");
        return;
      }

      const reconciled = await reconcileAttempt(id);
      log(reconciled);
    } catch (error) {
      log(error.message || String(error));
    }
  }

  async function bootstrapFromURL() {
    try {
      const id = getAttemptIdFromURL();
      if (!id) {
        return;
      }

      attemptId = id;
      const reconciled = await reconcileAttempt(id);
      log(reconciled);
    } catch (error) {
      log(error.message || String(error));
    }
  }

  createForm.addEventListener("submit", handleCreateSubmit);
  confirmButton.addEventListener("click", handleConfirmClick);
  reconcileButton.addEventListener("click", handleReconcileClick);

  bootstrapFromURL();
})();
