import React from "react";
import { createRoot, Root } from "react-dom/client";

import App from "./App/App";
import { ID_UNSET } from "./App/lib/api";

const rootElement = document.getElementById("root");

function render(root: Root, id: number) {
    root.render(
        <React.StrictMode>
            <App id={id} setId={(newId: number) => {
                id = newId;
                render(root, id);
            }}/>
        </React.StrictMode>
    );
}

if (rootElement) {
    const root = createRoot(rootElement);
    let id = ID_UNSET;
    render(root, id);
}
else {
    throw new Error("No root element!!!");
}
