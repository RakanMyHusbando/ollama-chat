const modelInput = document.querySelector(".model.form input");
const modelButton = document.querySelector(".model.form button");
import { Ollama } from "./script.js";
const ollama = new Ollama();

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
            const res = JSON.parse(str);
            const p = document.createElement("p");
            p.textContent = `status: ${res.status ? res.status : "-"} [${res.completed ? res.completed : "-"}/${res.total ? res.total : "-"}]`;
            document.querySelector("main").appendChild(p);
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
