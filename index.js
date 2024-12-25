const ollama = require("ollama");

ollama
    .chat({
        model: "codellama:13b",
        message: [
            {
                role: "user",
                content: "Hello, how are you?",
            },
        ],
    })
    .then((resp) => {
        console.log(resp.message.content);
    });
