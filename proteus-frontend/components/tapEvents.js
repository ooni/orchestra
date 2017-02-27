import injectTapEventPlugin from 'react-tap-event-plugin'
try {
  injectTapEventPlugin()
} catch (e) {
  // Do nothing, just preventing error
}
