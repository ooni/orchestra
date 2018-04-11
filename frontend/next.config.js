const webpack = require('webpack')

if (process.env.NODE_ENV !== 'production') {
  require('dotenv').config();
}

module.exports = {
  webpack: (config) => {
    if (config.resolve.alias) {
      // We need to remove the react-dom alias because it breaks react-tap-event-plugin in production
      delete config.resolve.alias['react-dom']
      delete config.resolve.alias['react']
    }
    config.plugins.push(
      new webpack.DefinePlugin({
        'process.env.REGISTRY_URL': JSON.stringify(process.env.REGISTRY_URL),
        'process.env.ORCHESTRATE_URL': JSON.stringify(process.env.ORCHESTRATE_URL)
      })
    )
    return config
  }
}
