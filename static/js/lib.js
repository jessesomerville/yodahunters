function formatLocalTimestamps() {
    const timeOpts = { year: 'numeric', month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit', second: '2-digit' };
    const dateOpts = { year: 'numeric', month: 'short', day: 'numeric' };

    document.querySelectorAll('.threadbox-comment-ts, .threadbox-lastpost-ts').forEach(el => {
        const d = new Date(el.textContent.trim());
        if (!isNaN(d)) el.textContent = d.toLocaleString(undefined, timeOpts);
    });

    document.querySelectorAll('.user-view-regdate').forEach(el => {
        const match = el.textContent.match(/(.*)(\d{4}-\d{2}-\d{2}T.+)/);
        if (!match) return;
        const d = new Date(match[2].trim());
        if (!isNaN(d)) el.textContent = match[1] + d.toLocaleDateString(undefined, dateOpts);
    });
}

document.addEventListener('DOMContentLoaded', formatLocalTimestamps);

function jsonPost(path, data, error, redir = null) {
    options = {
        method: "POST",
        headers: {
        Accept: "application/json, text/plain, */*",
        "Content-Type": "application/json",
        },
        body: JSON.stringify(data)
    }
    return fetch(path, options)
        .then(response => {
            if (!response.ok) {
                alert(error)
                throw new Error(`HTTP error! status: ${response.status}`);
            } else {
                if (redir != null) {
                    window.location.href = redir
                }
                return response.json();
            }
        })
        .then(data => {
            return data;
        })
        .catch(error => {
            console.error('Error:', error); // Handle any errors during the fetch operation
    });
}