window.addEventListener('load', (event) => {
    makeDatasetRowsClickable();
    makeRecordRowsClickable();
});

function makeDatasetRowsClickable() {
    document.querySelectorAll('.datasetRow').forEach(row => {
        row.addEventListener('click', event => {
            window.location = row.dataset.view
        });
    })
}

function makeRecordRowsClickable() {
    document.querySelectorAll('.recordRow').forEach(row => {
        row.addEventListener('click', event => {
            window.location = row.dataset.view
        });
    })
}