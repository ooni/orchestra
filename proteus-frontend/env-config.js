const prod = process.env.NODE_ENV === 'production'
module.exports = {
  'REGISTRY_URL': prod ? 'https://registry.proteus.ooni.io' : 'http://localhost:8080',
  'EVENTS_URL': prod ? 'https://events.proteus.ooni.io' : 'http://localhost:8082'
}
