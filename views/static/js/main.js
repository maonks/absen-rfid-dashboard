// Script Sidebar ==================================================================================

  function toggleSidebar() {
    const sidebar = document.getElementById("sidebar");
    const overlay = document.getElementById("overlay");

    sidebar.classList.toggle("-translate-x-full");
    overlay.classList.toggle("hidden");
  }


//Script Websocket ==================================================================================

  // const ws = new WebSocket("ws://192.168.1.100:8000/websocket");
  const ws = new WebSocket("wss://absenrfid.mainsambilbelajar.com/websocket");
  ws.onmessage = (e) => {

    let data;
    try {
      data = JSON.parse(e.data);
    } catch {
      return; // abaikan non-json
    }

    // ===============================
    // 1️⃣ EVENT ABSENSI (DEFAULT)
    // ===============================
    if (data.uid) {

      /* ===============================
         A. REALTIME TABLE (monitoring)
      =============================== */
      const absenTable = document.getElementById("absen-table");
      if (absenTable) {
        htmx.trigger(absenTable, "reload");
      }

      /* ===============================
         B. HOME DASHBOARD (hari ini)
      =============================== */
      const row = document.getElementById("row-" + data.uid);

      if (row) {
        // siswa sudah ada → refresh baris saja
        htmx.trigger(row, "refresh");
      } else {
        // siswa baru → reload wrapper
        const home = document.getElementById("home-realtime");
        if (home) {
          htmx.trigger(home, "refresh");
        }
      }

      /* ===============================
         C. ABSENSI BULANAN (BARU)
      =============================== */
      const bulanan = document.getElementById("absensi-bulanan");
      if (bulanan) {
        htmx.trigger(bulanan, "reload");
      }
    }

    // ===============================
    // 2️⃣ EVENT LAIN (DI MASA DEPAN)
    // ===============================
    // if (data.type === "device") {}
    // if (data.type === "logout") {}

  };


//Script Close modal  ==================================================================================


  function closeModal() {
    const modal = document.getElementById("modal");
    if (modal) modal.innerHTML = "";
  }


// Script Logout  ======================================================================================

  function logout() {
    fetch("/logout", {
      method: "POST",
      credentials: "include" // ⬅️ WAJIB
    }).then(() => {
      window.location.href = "/login";
    });
  }

//Script PWA Service worker  ============================================================================

  console.log("🟡 main layout loaded");

  if ("serviceWorker" in navigator) {
    console.log("🟡 SW supported");

    window.addEventListener("load", () => {
      console.log("🟡 window load");

      navigator.serviceWorker
        .register("/pwa/sw.js")
        .then(reg => {
          console.log("✅ SW REGISTERED:", reg.scope);
        })
        .catch(err => {
          console.error("❌ SW REGISTER FAILED:", err);
        });
    });
  } else {
    console.warn("❌ Service Worker NOT supported");
  }
