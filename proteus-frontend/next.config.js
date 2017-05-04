const webpack = require('webpack')

if (process.env.NODE_ENV !== 'production') {
  require('dotenv').config();
}

module.exports = {
  webpack: (config) => {
    config.plugins.push(
      new webpack.DefinePlugin({
        'process.env.REGISTRY_URL': JSON.stringify(process.env.REGISTRY_URL),
        'process.env.EVENTS_URL': JSON.stringify(process.env.EVENTS_URL)
      })
    )
    return config
  }
}
