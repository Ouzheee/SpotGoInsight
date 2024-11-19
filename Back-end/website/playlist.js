// Replace these values with your own
const accessToken = 'BQAM-JSdJLQPpEmptcpNSIpOPIjaxI8t83OMqdH50KX3rykibg8dXxhor8zlcGDa2_zjt4JdnJV1ioEbRfut9ipIN-J9cYN8nWUWp-PnpDMQWYcUaiU'; // OAuth Token
const playlistId = '37i9dQZF1E35SYJigVST3q'; // Spotify Playlist ID
const apiUrl = `https://api.spotify.com/v1/playlists/37i9dQZF1E35SYJigVST3q?si=ad5eccaa5a364b9c/tracks`;

document.getElementById('loadPlaylist').addEventListener('click', () => {
  fetchPlaylist();
});

function fetchPlaylist() {
  fetch(apiUrl, {
    headers: {
      'Authorization': `Bearer ${accessToken}`,
    },
  })
    .then((response) => {
      if (!response.ok) throw new Error('Failed to fetch playlist');
      return response.json();
    })
    .then((data) => {
      displayPlaylist(data);
    })
    .catch((error) => {
      console.error('Error:', error);
      alert('Failed to load playlist. Check the console for details.');
    });
}

function displayPlaylist(playlist) {
  const playlistContainer = document.getElementById('playlist');
  playlistContainer.innerHTML = `<h2>"Daily Mix 6"</h2>`;

  playlist.tracks.items.forEach((item) => {
    const track = item.track;
    const trackElement = document.createElement('div');
    trackElement.classList.add('track');
    trackElement.innerHTML = `
      <img src="${track.album.images[0]?.url || ''}" alt="Album Art">
      <div>
        <strong>${track.name}</strong> by ${track.artists.map((artist) => artist.name).join(', ')}
      </div>
    `;
    playlistContainer.appendChild(trackElement);
  });
}
