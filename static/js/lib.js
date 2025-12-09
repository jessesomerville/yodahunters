
function jsonPost(path, data, error, redir) {
    options = {
        method: "POST",
        headers: {
        Accept: "application/json, text/plain, */*",
        "Content-Type": "application/json",
        },
        body: JSON.stringify(data)
    }
    fetch(path, options)
        .then(response => {
            if (!response.ok) {
                alert(error)
                throw new Error(`HTTP error! status: ${response.status}`);
            } else {
                window.location.href = redir
            }
        })
        .catch(error => {
            console.error('Error:', error); // Handle any errors during the fetch operation
    });
}