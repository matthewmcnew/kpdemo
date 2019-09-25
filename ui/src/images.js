export async function fetchImages() {
    const response = await fetch('/images');
    const body = await response.json();
    if (response.status !== 200) throw Error(body.message);

    body.sort(sort);

    return body;
}


function sort(a, b) {
    const aCreatedAt = Date.parse(a.createdAt);
    const bCreatedAt = Date.parse(b.createdAt);

    if (aCreatedAt === bCreatedAt) {
        return nameSort(a, b);
    }

    if (aCreatedAt < bCreatedAt) {
        return -1;
    }
    if (aCreatedAt > bCreatedAt) {
        return 1;
    }
    return 0;
}

function nameSort(a, b) {
    if (a.name < b.name) {
        return -1;
    }
    if (a.name > b.name) {
        return 1;
    }
    return 0;
}

