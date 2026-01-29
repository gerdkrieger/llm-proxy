console.log('✅ main-test.js wird geladen');

// Warte bis DOM geladen ist
document.addEventListener('DOMContentLoaded', () => {
  console.log('✅ DOM ist bereit');
  
  const app = document.getElementById('app');
  
  if (!app) {
    console.error('❌ #app Element nicht gefunden!');
    return;
  }
  
  console.log('✅ #app Element gefunden');
  
  // Erstelle einfaches HTML (ohne Svelte)
  app.innerHTML = `
    <div style="
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      font-family: Arial, sans-serif;
    ">
      <div style="
        background: white;
        padding: 40px;
        border-radius: 10px;
        box-shadow: 0 10px 40px rgba(0,0,0,0.3);
        text-align: center;
        max-width: 500px;
      ">
        <h1 style="color: #667eea; margin: 0 0 20px 0;">
          🎉 JavaScript funktioniert!
        </h1>
        <p style="color: #666; margin-bottom: 20px;">
          Das bedeutet, Vite lädt die Dateien korrekt.
        </p>
        <div style="
          padding: 15px;
          background: #d4edda;
          border: 1px solid #c3e6cb;
          border-radius: 5px;
          color: #155724;
          font-weight: bold;
          margin-bottom: 20px;
        ">
          ✅ Vanilla JavaScript läuft
        </div>
        <button id="testBtn" style="
          background: #667eea;
          color: white;
          border: none;
          padding: 10px 30px;
          border-radius: 5px;
          cursor: pointer;
          font-size: 16px;
        ">
          Test Button
        </button>
        <div id="result" style="margin-top: 20px;"></div>
      </div>
    </div>
  `;
  
  console.log('✅ HTML wurde in #app eingefügt');
  
  // Test Button Event
  const btn = document.getElementById('testBtn');
  if (btn) {
    btn.addEventListener('click', () => {
      console.log('✅ Button Click funktioniert');
      document.getElementById('result').innerHTML = `
        <div style="
          padding: 10px;
          background: #d1ecf1;
          border: 1px solid #bee5eb;
          border-radius: 5px;
          color: #0c5460;
          margin-top: 10px;
        ">
          ✅ Event Listener funktioniert!<br>
          <small>Svelte ist das Problem, nicht JavaScript.</small>
        </div>
      `;
    });
  }
});

console.log('✅ main-test.js komplett geladen');
