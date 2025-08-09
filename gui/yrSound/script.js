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
const playSong = async(songDiv, id) => {
	const cover = document.querySelector("#player > .simg-cont > img");
	const title = document.querySelector("#player > .song-info > b");
	const artist = document.querySelector("#player > .song-info > span");
	const songBar = document.getElementById("song-bar");

	fetch(getFetchBall("get-song-info", `id=${id}`)).then(response => {
		if (response.ok) return response.json();
	}).then(response => {
		title.textContent = response.title;
		artist.textContent = response.artist;
	});
	fetch(getFetchBall("get-song-blob", `id=${id}`)).then(response => {
		if (response.ok) return response.blob();
	}).then(response => {
		document.getElementById("song").src = URL.createObjectURL(response);
		document.getElementById("song").load();
		
		const onload = () => {
			songBar.min = "0"; songBar.value = "0";
			songBar.max = `${document.getElementById("song").duration}`;
			document.getElementById("song").play();
			document.getElementById("song").removeEventListener("loadedmetadata", onload);
		}
		document.getElementById("song").addEventListener("loadedmetadata", onload);
	});
	fetch(getFetchBall("get-cover", `id=${id}`)).then(response => {
		if (response.ok) return response.blob();
	}).then(response => {
		const rawIMG = new Image();
		rawIMG.src = cover.src = URL.createObjectURL(response);

		rawIMG.onload = () => {
			cover.style.aspectRatio = `${rawIMG.naturalWidth} / ${rawIMG.naturalHeight}`;
			cover.style.width = "auto";
		}
	}).catch(() => {
		cover.src = "./vectors/song.svg";
		cover.style.aspectRatio = "1/1";
		cover.style.width = "32px";
	});

	currentSong = id;
	for (const element of Array.from(document.getElementById("songs").getElementsByClassName("song"))) {
		element.querySelector(".simg-cont > div").style.display = "none";
	}
	songDiv.querySelector(".simg-cont > div").style.display = "flex";
}

let albumsClickable = true;
let songsClickable = true;
let artistsClickable = true;
let currentSong;
let queue = [];
const showAllAlbums = async() => {	
	if (!albumsClickable) return;
	for (const element of Array.from(document.getElementById("albums").childNodes)) element.remove();
	
	fetch(getFetchBall("get-albums"))
	.then(response => {
		if (response.ok) return response.json();
	}).then(albums => {
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
				}).catch(() => {
					albumIMG.src = "./vectors/song.svg";
					albumIMG.style.aspectRatio = "1/1";
					albumIMG.style.width = "64px";
					albumDiv.appendChild(albumIMG);
				});
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

const showAllArtists = async() => {
	if (!artistsClickable) return;
	for (const element of Array.from(document.getElementById("artists").childNodes)) element.remove();
	
	fetch(getFetchBall("get-artists"))
	.then(response => {
		if (response.ok) return response.json();
	}).then(artists => {
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

const showAllSongs = async() => {
	if (!songsClickable) return;
	fetch(getFetchBall("get-songs"))
	.then(response => {
		if (response.ok) return response.json();
	}).then(songs => {
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
			songDiv.dataset.id = song.id;
			content.appendChild(songDiv);
			songDiv.addEventListener("click", () => playSong(songDiv, song.id));

			const songIMGCont = document.createElement("div");
			songIMGCont.classList.add("simg-cont");
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
			}).catch(() => {
				songIMG.src = "./vectors/song.svg";
				songIMG.style.aspectRatio = "1/1";
				songIMG.style.width = "20px";
				songIMGCont.appendChild(songIMG);
			});
			const playingIcon = document.createElement("div");
			songIMGCont.appendChild(playingIcon);
			const piIMG = document.createElement("img");
			piIMG.src = "./vectors/playing.svg";
			piIMG.width = "20"; playingIcon.appendChild(piIMG);
			playingIcon.style.display = songDiv.dataset.id === currentSong? "flex":"none";			
		
			const songInfo = document.createElement("div");
			songInfo.classList.add("sinfo");
			songDiv.appendChild(songInfo);
		
			const songTitle = document.createElement("b");
			songTitle.textContent = song.title;
			songInfo.appendChild(songTitle);

			const songArtist = document.createElement("span");
			songArtist.textContent = song.artist;
			songInfo.appendChild(songArtist);
		}
	});
} 

let sidebaropened = false;
const sidebar = document.getElementById("sidebar");
const button = document.getElementById("sidebar-button");
const icon = button.querySelector("img");

const hideSidebar = () => {
	sidebaropened = false;
	sidebar.style.transform = "";
	icon.src = "./vectors/hamburger-menu.svg";
	button.style.top = "";
	setTimeout(() => {
		sidebar.style.display = "";
	}, 400);
}
const showSidebar = () => {
	sidebaropened = true;
	sidebar.style.display = "flex";
	sidebar.style.transform = "translateY(-100%)";
	requestAnimationFrame(() => {
		requestAnimationFrame(() => {
			sidebar.style.transform = "translateY(0)";
		});
	});

	icon.src = "./vectors/close.svg";
	button.style.top = "360px";
}
button.addEventListener("click", () => sidebaropened? hideSidebar():showSidebar());
document.getElementById("show-songs").addEventListener("click", () => {
	hideSidebar();
	showAllSongs();
});
document.getElementById("show-artists").addEventListener("click", () => {
	hideSidebar();
	showAllArtists();
});
document.getElementById("show-albums").addEventListener("click", () => {
	hideSidebar();
	showAllAlbums();
});
document.getElementById("song").addEventListener("loadedmetadata", () => {
	const songBar = document.getElementById("song-bar");
	songBar.max = document.getElementById("song").duration;
});

document.getElementById("song-bar").addEventListener("input", () => {
	const songBar = document.getElementById("song-bar");
	document.getElementById("song").currentTime = songBar.value;
});

document.getElementById("song").addEventListener("timeupdate", () => {
	const songBar = document.getElementById("song-bar");
	songBar.value = document.getElementById("song").currentTime;

	const elapsed = document.querySelector("#player > .song-info > .time > .elapsed");
	const minutes = Math.floor(document.getElementById("song").currentTime / 60);
	const seconds = Math.floor(document.getElementById("song").currentTime % 60).toString().padStart(2, "0");
	elapsed.textContent = `${minutes}:${seconds}`;

	const total = document.querySelector("#player > .song-info > .time > .total");
	const minutes1 = Math.floor(document.getElementById("song").duration / 60);
	const seconds1 = Math.floor(document.getElementById("song").duration % 60).toString().padStart(2, "0");
	total.textContent = `${minutes1}:${seconds1}`;
});

document.getElementById("song").addEventListener("play", () => {
	const play = document.getElementById("play");
	play.querySelector("img").src = "./vectors/pause.svg";
});
document.getElementById("song").addEventListener("pause", () => {
	const play = document.getElementById("play");
	play.querySelector("img").src = "./vectors/play.svg";
});
document.getElementById("repeat").addEventListener("click", () => {
	if (document.getElementById("song").loop) {
		document.getElementById("song").loop = false;
		document.getElementById("repeat").querySelector("img").src = "./vectors/repeat.svg"
	} else {
		document.getElementById("song").loop = true;
		document.getElementById("repeat").querySelector("img").src = "./vectors/repeat-toggled.svg"
	}
});

document.getElementById("play").addEventListener("click", () => {
	document.getElementById("song").paused?
		document.getElementById("song").play():
		document.getElementById("song").pause();
});