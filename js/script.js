class Ollama {
    /** @type {string} */
    baseUrl = document.cookie.match(new RegExp(`(^| )ollama=([^;]+)`)).pop();

    /** @type {string[]} */
    models = [];

    /** @param {... string} apiPath */
    #url = (...apiPath) => `${this.baseUrl}/api/${apiPath.join("/")}`;

    getModel = async () => {
        const res = await fetch(this.#url("tags"));
        const data = await res.json();
        data.models.forEach((model) => this.models.push(model.name));
    };

    #httpError = (status, statusText) =>
        new Error(`HTTP Error! Status: ${status}: ${statusText}`);

    /**
     * @param {string} model
     * @param {{role: string, content: string}[]} messages
     * @returns {Promise<Response>}
     */
    chatStream = async (model, messages) => {
        try {
            const res = await fetch(this.#url("chat"), {
                method: "POST",
                body: JSON.stringify({ model, messages }),
            });
            if (!res.ok) throw this.#httpError(res.status, res.statusText);
            if (!res.body) throw new Error("Readable stream not found.");
            return res;
        } catch (error) {
            console.error(error);
        }
    };

    generate = async (model, prompt) => {
        try {
            const res = await fetch(this.#url("generate"), {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ model, prompt, stream: false }),
            });
            if (!res.ok) throw this.#httpError(res.status, res.statusText);
            return res.json();
        } catch (error) {
            console.error(error);
        }
    };

    /**
     * Pulls a stream of data from the server.
     * @param {string} model - The model to use for the stream.
     * @returns {Response} - The reader for the stream.
     */
    pullStream = async (model) => {
        try {
            const res = await fetch(this.#url(["pull"]), {
                method: "POST",
                body: JSON.stringify({ model }),
            });
            if (!res.ok) throw this.#httpError(res.status, res.statusText);
            if (!res.body) throw new Error("Readable stream not found.");
            return res;
        } catch (error) {
            console.error(error);
        }
    };
}

class Api {
    #httpError = (status, statusText) =>
        new Error(`HTTP Error! Status: ${status}: ${statusText}`);

    /**
     * @param {string} url
     * @param {string} body
     */
    post = async (url, body) => {
        try {
            const res = await fetch(url, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: body,
            });
            if (!res.ok) throw this.#httpError(res.status, res.statusText);
            return;
        } catch (error) {
            console.error(error);
        }
    };

    #jsonToMessage = (msg) =>
        new Message(msg.chat_id, msg.role, msg.content, msg.created_at);

    #jsonToChat = (chat) => {
        return new Chat(
            chat.id,
            chat.messages
                ? chat.messages.map((msg) => this.#jsonToMessage(msg))
                : [],
            chat.created_at,
            chat.name,
        );
    };

    /**
     * @param {string|null} chatId
     * @param {true} [loadMessages]
     * @returns {Promise<Chat[]>}
     */
    getChats = async (chatId, loadMessages) => {
        try {
            const result = [];
            let query = chatId ? "?id=" + chatId : "";
            query += loadMessages ? (query != "" ? "&" : "?") + "msg=true" : "";
            const res = await fetch(`/api/chat${query}`);
            if (!res.ok) throw this.#httpError(res.status, res.statusText);
            const data = await res.json();
            data.forEach((chat) => result.push(this.#jsonToChat(chat)));
            return result;
        } catch (error) {
            console.error(error);
        }
    };

    /**
     * @param {string} url
     * @param {string} body
     */
    update = async (url, body) => {
        try {
            const res = await fetch(url, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body,
            });
            if (!res.ok) throw this.#httpError(res.status, res.statusText);
            return;
        } catch (error) {
            console.error(error);
        }
    };
}

class Chat extends Ollama {
    /** @type {{id: string, name: string, createdAt: string, messages: Message[] }} */
    content;
    /** @type {Api} */
    api;

    /**
     * @param {string} [id]
     * @param {Message[]} [messages]
     * @param {string} [createdAt]
     * @param {string} [name]
     */
    constructor(id, messages, createdAt, name) {
        super();
        this.content = {
            id: id ? id : this.#makeId(),
            name: name ? name : "-",
            createdAt: createdAt ? createdAt : new Date().toISOString(),
            messages: messages ? messages : [],
        };
        this.api = new Api();
    }

    #makeId = () => Math.floor(Math.random() * Date.now()).toString(36);

    updateName = async () => {
        const task =
            "Create a name for follwoing chat with maximum length of 3 words. Only respond with the name.";
        let chat = "";
        this.content.messages.forEach((msg) => {
            chat += `\t${msg.content.role}: ${
                msg.content.role == "assistant" &&
                msg.content.content.includes("[/think]")
                    ? msg.content.content.split("[/think]")[1]
                    : msg.content.content
            }\n`;
        });
        await this.generate(
            this.models[0],
            `<task>\n\t${task}\n</task>\n<chat>\n${chat}</chat>`,
        )
            .then((res) => {
                console.log(res);
                this.content.name = res.response.includes("</think>")
                    ? res.response.split("</think>")[1]
                    : res;
            })
            .then(
                async () =>
                    await this.api.update(
                        "/api/chat",
                        JSON.stringify({
                            name: this.content.name,
                            id: this.content.id,
                        }),
                    ),
            );
    };

    post = async () =>
        this.api.post("/api/chat", JSON.stringify(this.formJson()));

    /**
     * @param {string} role - message role (user/assistant)
     * @param {string} content - message content
     * @param {string} [createdAt] - message creation time
     * @returns {Message} - The message object.
     */
    addMessage = (role, content, createdAt = undefined) => {
        const msg = new Message(this.content.id, role, content, createdAt);
        this.content.messages.push(msg);
        return msg;
    };

    /**
     * @param  {Uint8Array<ArrayBufferLike> | undefined} value
     * @returns {string}
     */
    #messageContent = (value) => {
        let text = "";
        if (value)
            new TextDecoder()
                .decode(value)
                .split("}\n")
                .forEach((e) => {
                    if (e != "") text += JSON.parse(e + "}").message.content;
                });
        return text;
    };

    /**
     * @param {Message} message
     * @param {string} model
     */
    addOllamaMessage = async (message, model) => {
        console.log(this.formJson().messages);
        const reader = await this.chatStream(
            model,
            this.formJson().messages,
        ).then((res) => res.body.getReader());
        let check = false;
        do {
            let { done, value } = await reader.read();
            check = done;
            if (done)
                return message
                    .post()
                    .then(() => this.content.messages.push(message));
            const text = this.#messageContent(value);
            message.addText(text);
        } while (!check);
    };

    /** @returns {Object} The JSON representation of the chat object.*/
    formJson = () => {
        return {
            id: this.content.id,
            name: this.content.name,
            created_at: this.content.createdAt,
            messages: this.content.messages.map((msg) => msg.formJson()),
        };
    };
}

class Message {
    /** @type {{ chatId: string, content: string, role: string, createdAt: string }} */
    content;
    /** @type {HTMLDivElement} */
    htmlElement;
    /** @type {Api} */
    api;

    /**
     * @param {string} chatId
     * @param {string} role
     * @param {string} [content]
     * @param {string} [createdAt]
     */
    constructor(chatId, role, content, createdAt) {
        this.content = {
            chatId,
            content: content ? content : "",
            role,
            createdAt: createdAt ? createdAt : new Date().toISOString(),
        };
        this.api = new Api();
    }

    post = async () =>
        await this.api.post("/api/message", JSON.stringify(this.formJson()));

    createHTML = () => {
        this.htmlElement = document.createElement("div");
        const msgTextElem = document.createElement("div");
        msgTextElem.classList.add("text");
        msgTextElem.innerText = this.content.content;
        this.htmlElement.classList.add("message", this.content.role);
        this.htmlElement.appendChild(msgTextElem);
        return this.htmlElement;
    };

    /** @param {string} text */
    addText = (text) => {
        let t = text
            .replace("<think>", "[think]")
            .replace("</think>", "[/think]");
        this.content.content += t;
        if (this.htmlElement)
            this.htmlElement.firstElementChild.textContent += t;
    };

    /** @returns {Object} The JSON representation of the message object.*/
    formJson = () => {
        return {
            chat_id: this.content.chatId,
            content: this.content.content,
            role: this.content.role,
            created_at: this.content.createdAt,
        };
    };
}

export { Api, Chat, Message, Ollama };
