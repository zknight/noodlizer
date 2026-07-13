function togviz(id, kind) {
    var el = document.getElementById(id);
    if (el.style.display == 'none') {
        el.style.display = kind;
    } else {
        el.style.display = 'none';
    }
}