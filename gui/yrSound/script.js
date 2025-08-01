// authentication
const go_on = (addAuth) => {
	if (addAuth) {
		username = usernameInput.value;
		password = passwordInput.value;
		use_auth = true;
	} else use_auth = false;
	document.getElementById("authentication").style.display = "none";
	document.getElementById("mainContent").style.display = "flex";
	showAllSongs();
};

let use_auth;
let username, password;
const authenticationDiv = document.getElementById("authentication");
const usernameInput = document.getElementById('username-input');
const passwordInput = document.getElementById('password-input');
const authenticateButton = document.getElementById('authenticateButton');
const dim_frame = document.getElementById("dim-frame");

fetch("/users-qm")
.then(response => {
	if (!response.ok) go_on(false);
	else authenticationDiv.style.display = "flex";
});

authenticateButton.addEventListener("click", () => {
	fetch(`/isUser-qm?username=${encodeURIComponent(usernameInput.value)}&password=${encodeURIComponent(passwordInput.value)}`)
	.then(response => {
		if (response.ok) go_on(true);
		else {
			document.querySelector('#authentication > span').textContent = "Invalid login.";
			setTimeout(() => {
				document.querySelector('#authentication > span').textContent = "Login to see your music.";
			}, 1000);
		}
	})
});

// main
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

let albumsClickable = true;
let songsClickable = true;
let artistsClickable = true;
const showAllAlbums = () => {	
	for (const element of Array.from(document.getElementById("albums").childNodes)) element.remove();
	
	fetch(getFetchBall("get-albums"))
	.then(response => {
		if (response.ok) return response.json();
	}).then(albums => {
		if (!albumsClickable) return;
		albumsClickable = false;
		const content = document.getElementById("albums");
		content.style.display = "flex";

		document.getElementById("songs").style.display = "none";
		document.getElementById("artists").style.display = "none";
		songsClickable = artistsClickable = true;
		document.getElementById("show-songs").style.backgroundColor = "";
		document.getElementById("show-albums").style.backgroundColor = "#424242";
		document.getElementById("show-artists").style.backgroundColor = "";

		for (const album of albums) {
			const albumDiv = document.createElement("div");
			albumDiv.classList.add("album");
			content.appendChild(albumDiv);

			const albumIMG = document.createElement("img");
			fetch(getFetchBall("get-album"), {method: "POST", body: JSON.stringify({
				album: album.title, "album-artist": album.artist
			})}).then(response => {
				if (response.ok) return response.json();
			}).then(response => {
				fetch(getFetchBall(`get-cover`, `id=${response[0].id}`)).then(response => {
					if (response.ok) return response.blob();
				}).then(response => {
					const rawIMG = new Image();
					albumIMG.src = rawIMG.src = URL.createObjectURL(response);

					rawIMG.onload = () => {
						albumIMG.style.aspectRatio = `${rawIMG.naturalWidth} / ${rawIMG.naturalHeight}`
						albumDiv.appendChild(albumIMG);
					}
				})
			});
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

const showAllArtists = () => {
	for (const element of Array.from(document.getElementById("artists").childNodes)) element.remove();
	
	fetch(getFetchBall("get-artists"))
	.then(response => {
		if (response.ok) return response.json();
	}).then(artists => {
		if (!artistsClickable) return;
		artistsClickable = false;
		const content = document.getElementById("artists");
		content.style.display = "flex";

		document.getElementById("songs").style.display = "none";
		document.getElementById("albums").style.display = "none";
		albumsClickable = songsClickable = true;
		document.getElementById("show-songs").style.backgroundColor = "";
		document.getElementById("show-albums").style.backgroundColor = "";
		document.getElementById("show-artists").style.backgroundColor = "#424242";

		for (const artist of artists) {
			const artistDiv = document.createElement("div");
			artistDiv.classList.add("artist");
			content.appendChild(artistDiv);

			const picCont = document.createElement("div");
			artistDiv.appendChild(picCont);

			const img = document.createElement("img");
			img.src = "../../vectors/artist.svg";
			picCont.appendChild(img);
/* 			fetch(getFetchBall("get-artist-picture"), {method: "POST", body: JSON.stringify({name: artist})})
			.then(response => {
				if (response.ok) return response.blob();
			}).then(image => {
				if (typeof image !== "undefined" && image.length !== 0) {
					console.log("available");
					const rawIMG = new Image();
					img.src = rawIMG.src = URL.createObjectURL(image);
					rawIMG.onload = () => {
						img.style.aspectRatio = `${rawIMG.naturalWidth} / ${rawIMG.naturalHeight}`;
						picCont.appendChild(img);
					}
				} else {
					console.log("not available");
					img.src = "../../vectors/artist.svg";
					img.style.aspectRatio = `1/1`;
					picCont.appendChild(img);
				}
			}); */

			const artistSpan = document.createElement("span");
			artistSpan.textContent = artist;
			artistDiv.appendChild(artistSpan);
		}
	})
}

const showAllSongs = () => {
	for (const element of Array.from(document.getElementById("songs").childNodes)) element.remove();

	fetch(getFetchBall("get-songs"))
	.then(response => {
		if (response.ok) return response.json();
	}).then(songs => {
		if (!songsClickable) return;
		songsClickable = false;
		const content = document.getElementById("songs");
		content.style.display = "flex";
	
		document.getElementById("artists").style.display = "none";
		document.getElementById("albums").style.display = "none";
		albumsClickable = artistsClickable = true;
		document.getElementById("show-songs").style.backgroundColor = "#424242";
		document.getElementById("show-albums").style.backgroundColor = "";
		document.getElementById("show-artists").style.backgroundColor = "";

		for (const song of songs) {
			const songDiv = document.createElement("div");
			songDiv.classList.add("song");
			content.appendChild(songDiv);
			
			const songIMGCont = document.createElement("div");
			songDiv.appendChild(songIMGCont);

			const songIMG = document.createElement("img");
			fetch(getFetchBall("get-cover", `id=${song.id}`)).then(response => {
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

document.getElementById("show-songs").addEventListener("click", showAllSongs);
document.getElementById("show-artists").addEventListener("click", showAllArtists);
document.getElementById("show-albums").addEventListener("click", showAllAlbums);
