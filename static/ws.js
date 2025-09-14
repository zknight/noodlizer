onload = (evt) => {
    let client_id
    function dial() {
        const conn = new WebSocket(`ws://${location.host}/subscribe`)

        conn.addEventListener('close', ev => {
            if (ev.code != 1001) {
                console.info(`ev.code = ${ev.code} Reconnect in 1s`)
                setTimeout(dial, 1000)
            }
        })
        conn.addEventListener('open', ev => {
            console.info('websocket connected')
        })
        conn.addEventListener('message', ev => {
            console.info(`received: ${ev.data}`)
            const msg = JSON.parse(ev.data)
            switch (msg.type) {
                case "sub":
                    client_id = msg.id
                    console.info(`got id: ${msg.id}`)
                    break
                case "wait":
                    showOverlay()
                    break
                case "proceed":
                    hideOverlay()
                    break
                default:
                    console.info("unexpected msg on ws: ", ev.data)
            }
        })
    }
    dial()

    const overlay = document.getElementById('overlay')
    const wait_btn = document.getElementById('wait')
    const ready_btn = document.getElementById('ready')
    const syncform = document.getElementById('syncform')

    ready_btn.disabled = true
    wait_btn.disabled = false

    function showOverlay() { overlay.style.display = 'block'; } //wait_btn.disabled = true; ready_btn.disabled = false; }
    function hideOverlay() { overlay.style.display = 'none'; } //wait_btn.disabled = false; ready_btn.disabled = true; }

    syncform.onsubmit = async ev => {
        ev.preventDefault()
       
        if (ev.submitter.name == 'wait') {
            wait_btn.disabled = true
            ready_btn.disabled = false
            // TODO send a request to wait all the subscribers
        } else {
            wait_btn.disabled = false
            ready_btn.disabled = true
            // TODO send a request to ready all the subscribers
        }
        console.info(ev.submitter.name)
        try {
            const resp = await fetch('/wait', {
                method: 'POST',
                body: `${client_id}=${ev.submitter.name}`
            })
            if (resp.status != 202) {
                throw new Error(`Unexpected HTTP Status ${resp.status} ${resp.statusText}`)
            }
        } catch (err) {
            console.error(`update wait failed: ${err.message}`)
        }
    }

}
//})()