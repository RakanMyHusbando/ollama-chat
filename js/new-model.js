const modelInput = document.querySelector(".model.form input");
const modelButton = document.querySelector(".model.form button");
import { Ollama } from "./script.js";
const ollama = new Ollama();

const createHTML = (text) => {
    const msgElem = document.createElement("div");
    msgElem.classList.add("message", "pull-model");
    const msgTextElem = document.createElement("div");
    msgTextElem.classList.add("text");
    msgTextElem.textContent = text;
    msgElem.appendChild(msgTextElem);
    return msgElem;
};

/**
 * @param {any} res
 * @returns {string}
 */
const formRespMsg = (res) => {
    const completed =
        res.completed && parseInt(res.completed) ? parseInt(res.completed) : 0;
    const total = res.total && parseInt(res.total) ? parseInt(res.total) : 1;
    const status = res.status ? res.status : "unknown";
    const pr = status.includes("pulling")
        ? ` \t${Math.round((100 * completed) / total)}%`
        : "";
    return `${status}${pr}`;
};

const pullModel = async () => {
    const reader = await ollama
        .pullStream(modelInput.value)
        .then((res) => res.body.getReader());
    let check = false;
    do {
        let { done, value } = await reader.read();
        check = done;

        const str = new TextDecoder().decode(value);
        try {
            /** @type {HTMLDivElement|null} */
            const main = document.querySelector("main");
            str.split("}\n").forEach((elem) => {
                if (elem != "") {
                    const res = JSON.parse(elem + "}");
                    if (res.status.includes("pulling")) main.innerHTML = "";
                    main.appendChild(createHTML(formRespMsg(res)));
                }
            });
        } catch (err) {
            console.log(err);
        }
        console.log(str);
    } while (!check);
};

modelButton.addEventListener("click", async () => await pullModel());
modelInput.addEventListener("keydown", async (e) => {
    if (e.key === "Enter") await pullModel();
});
