const express = require('express')
const next = require('next')

process.env.NODE_ENV = process.env.NODE_ENV || 'production'
process.env.PORT = process.env.PORT || 3000

const dev = process.env.NODE_ENV !== 'production'
if (dev === true) {
  process.env.REGISTRY_URL = process.env.REGISTRY_URL || "http://127.0.0.1:8080"
  process.env.ORCHESTRATE_URL = process.env.ORCHESTRATE_URL || "http://127.0.0.1:8082"
}

const app = next({ dir: '.', dev })
const handle = app.getRequestHandler()

app.prepare()
.then(() => {
  const server = express()

  server.all('*', (req, res) => {
    return handle(req, res)
  })

  server.listen(process.env.PORT, err => {
    if (err) {
      throw err
    }
    console.log('> Ready on http://localhost:' +
                process.env.PORT +
                '[' + process.env.NODE_ENV + ']')
  })
})
.catch(err => {
  console.log('An error occurred, unable to start the server')
  console.log(err)
})
