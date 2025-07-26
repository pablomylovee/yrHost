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

const showAllSongs = () => {
	fetch(getFetchBall("get-songs"))
	.then(response => {
		if (response.ok) return response.json();
	}).then(songs => {
		const content = document.getElementById("content");
		content.style.flexDirection = "column";
		content.style.flexWrap = "nowrap";
		content.style.alignItems = "center";
		content.style.justifyContent = "";

		for (const song of songs) {
			const songDiv = document.createElement("div");
			songDiv.classList.add("song");
			content.appendChild(songDiv);
			
			const songIMGCont = document.createElement("div");
			songDiv.appendChild(songIMGCont);

			const songIMG = document.createElement("img");
			const body = JSON.stringify({
				title: song.title, artist: song.artist,
				album: song.album, "album-artist": song["album-artist"],
			});
			fetch(getFetchBall("get-cover"), {method: "POST", body: body, headers: {
				"Content-Type": "application/json",
			}}).then(response => {
				if (response.ok) return response.blob();
			}).then(image => {
				const rawIMG = new Image();
				songIMG.src = rawIMG.src = URL.createObjectURL(image);

				rawIMG.onload = () => {
					songIMG.style.aspectRatio = `${rawIMG.naturalWidth} / ${rawIMG.naturalHeight}`;
					songIMGCont.appendChild(songIMG);
				}
			});
		
			const songSpan = document.createElement("span");
			songSpan.style.display = "flex";
			songSpan.style.gap = "4px";
			songSpan.innerHTML = `${song.title}<span style="color: #999;">•</span>${song.artist}`
			songDiv.appendChild(songSpan);
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
			showAllSongs();
        }
	});
});

document.getElementById("show-albums").addEventListener("click", () => {
    window.location.href = `../albums?username=${username}&password=${password}`;
});
