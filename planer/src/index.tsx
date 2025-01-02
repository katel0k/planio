import React from "react";
import { createRoot } from "react-dom/client";

import App from "./App";

const rootElement = document.getElementById("root");

if (rootElement) {
    const root = createRoot(rootElement);
    fetch("http://0.0.0.0:5000/join/artem")
        .then(() => {
            root.render(
                <React.StrictMode>
                    <App />
                </React.StrictMode>
            );
        })
}
else {
    throw new Error("No root element!!!");
}
