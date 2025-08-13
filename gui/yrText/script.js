const getFetchBall = (ballModel, notes) => {
	let resBall = `/${ballModel}`;
	resBall = `${resBall}?username=${params.get("username")}&password=${params.get("password")}`;
	if (typeof notes !== 'undefined') resBall = `${resBall}&${notes}`;

	return resBall;
}

const shortenName = (s, maxLength) => {
	if (s.length > maxLength) return s.slice(0, maxLength)+"...";
	else return s;	
}

const showNoti = (content) => {
    const notificationSpan = document.getElementById("notification");
    notificationSpan.style.display = "flex";
    setTimeout(() => notificationSpan.style.opacity = "1", 100);
    notificationSpan.textContent = content.toString();

    setTimeout(() => {
        notificationSpan.style.opacity = "0";
        setTimeout(() => notificationSpan.style.display = "", 400);
    }, 3500);
}

const params = new URLSearchParams(window.location.search);
const editable = (response) => {
    if (!response.ok) return false;

    const contentTypes = [
        "application/json",
        "application/javascript",
        "application/xml",
        "application/xhtml+xml",
        "application/ld+json",
        "application/graphql",
        "application/rss+xml",
        "application/atom+xml",
        "application/sql"
    ];

    const contentType = response.headers.get("Content-Type");
    if (!contentType) return false;

    const mimeType = contentType.split(";")[0].trim();

    if (!(mimeType.startsWith("text/") || contentTypes.includes(mimeType))) {
        return false;
    }

    return true;
};

const textarea = document.getElementById("writer");

fetch(getFetchBall(`yrFiles/files/${encodeURI(params.get("relative-path"))}`))
  .then(response => {
    if (!editable(response)) {
        document.getElementById("not-found").style.display = "flex";
        document.getElementById("bot").innerHTML = "";
        return null
    }
    return response.text();
  })
  .then(response => {
    if (typeof response !== "string") return;
    textarea.style.display = "block";
    textarea.disabled = false;
    textarea.value = response;

    const bot = document.getElementById("bot");
    bot.textContent = params.get("relative-path");
  }) 

document.addEventListener("keypress", (e) => {
    if (e.ctrlKey && e.key === "s")
        fetch(
            getFetchBall("write-to-file", `path=${encodeURI(params.get("relative-path"))}`),
            {method: "POST", body: textarea.value},
        ).then(response => {
            if (response.ok) showNoti(`Successfully wrote to ${params.get("relative-path")}`);
        });
});

