import React from "react";
import { createRoot } from "react-dom/client";

import App from "./App/App";

const rootElement = document.getElementById("root");

if (rootElement) {
    const root = createRoot(rootElement);
    root.render(
        <React.StrictMode>
            <App />
        </React.StrictMode>
    );
}
else {
    throw new Error("No root element!!!");
}
