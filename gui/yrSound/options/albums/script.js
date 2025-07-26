let username, password, use_auth;
const getFetchBall = (ballModel, notes) => {
	let resBall = `/${ballModel}`;
	if (use_auth) resBall = `${resBall}?username=${username}&password=${password}`;
	if (typeof notes !== 'undefined') {
		if (use_auth) resBall = `${resBall}&${notes}`;
		else resBall = `${resBall}?${notes}`;
	}

	return resBall;
}

const shortenName = (s, maxLength) => {
	if (s.length > maxLength) return s.slice(0, maxLength)+"...";
	else return s;	
}

const showAllAlbums = () => {
	for (const element of Array.from(document.getElementById("content").childNodes)) element.remove();
	
	fetch(getFetchBall("get-albums"))
	.then(response => {
		if (response.ok) return response.json();
	}).then(albums => {
		const content = document.getElementById("content");
		content.style.flexDirection = "row";
		content.style.flexWrap = "wrap";
		content.style.alignItems = "flex-start";
		content.style.justifyContent = "center";

		for (const album of albums) {
			const albumDiv = document.createElement("div");
			albumDiv.classList.add("album");
			content.appendChild(albumDiv);

			const albumIMG = document.createElement("img");
			const exSong = album.songs[0];
			const body = JSON.stringify({
				title: exSong.title, artist: exSong.artist,
				album: exSong.album, "album-artist": exSong["album-artist"],
			});
			fetch(getFetchBall("get-cover"), {method: "POST", body: body})
			.then(response => {
				if (response.ok) return response.blob();
			}).then(image => {
				const rawIMG = new Image();
				albumIMG.src = rawIMG.src = URL.createObjectURL(image);
				rawIMG.onload = () => {
					albumIMG.style.aspectRatio = `${rawIMG.naturalWidth} / ${rawIMG.naturalHeight}`
					albumDiv.appendChild(albumIMG);
				};
			})
			const albumInfoDiv = document.createElement("div");
			albumDiv.appendChild(albumInfoDiv);

			const aiH3 = document.createElement("h3");
			aiH3.textContent = shortenName(album.title, 20);
			albumInfoDiv.appendChild(aiH3);

			const aiSpan = document.createElement("span");
			aiSpan.textContent = `${shortenName(album.artist, 22)} • ${album.year}`;
			albumInfoDiv.appendChild(aiSpan);

			albumDiv.addEventListener("mouseenter", () => {
				albumInfoDiv.style.bottom = "0";
			});
			albumDiv.addEventListener("mouseleave", () => {
				albumInfoDiv.style.bottom = "-60px";
			});
		}
	})
}

const params = new URLSearchParams(window.location.search);
fetch("/users-qm").then(response => {
    if (!response.ok) {use_auth = false; showAllSongs();}
    else fetch(`/isUser-qm?username=${encodeURIComponent(params.get("username"))}&password=${encodeURIComponent(params.get("password"))}`)
	.then(response => {
		if (response.ok) {
			use_auth = true;
            username = params.get("username");
			password = params.get("password");
            document.body.style.display = "flex";
            showAllAlbums();
        }
	});
});

document.getElementById("show-songs").addEventListener("click", () => {
    window.location.href = `../songs?username=${username}&password=${password}`;
});
