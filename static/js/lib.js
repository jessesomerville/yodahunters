
function jsonPost(path, data, error, redir = None) {
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
                if (redir != None)
                    window.location.href = redir
                return response.json();
            }
        })
        .catch(error => {
            console.error('Error:', error); // Handle any errors during the fetch operation
    });
}