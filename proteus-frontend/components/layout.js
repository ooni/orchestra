import React from 'react'

import darkBaseTheme from 'material-ui/styles/baseThemes/darkBaseTheme'
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider'
import getMuiTheme from 'material-ui/styles/getMuiTheme'

import Home from 'material-ui/svg-icons/action/home'
import DeviceHub from 'material-ui/svg-icons/hardware/device-hub'
import Event from 'material-ui/svg-icons/action/event'

import Drawer from 'material-ui/Drawer'
import MenuItem from 'material-ui/MenuItem'
import RaisedButton from 'material-ui/RaisedButton'
import AppBar from 'material-ui/AppBar'

import Link from 'next/link'
import Head from 'next/head'

import './tapEvents'

export default class extends React.Component {

  static propTypes = {
    children: React.PropTypes.node.isRequired,
    title: React.PropTypes.string
  }

  constructor (props) {
    super(props)
    this.state = {
      drawerOpen: false
    }
    this.toggleDrawer = this.toggleDrawer.bind(this)
  }

  toggleDrawer () {
    this.setState({
      drawerOpen: !this.state.drawerOpen
    })
  }

  render () {
    const {
      drawerOpen
    } = this.state
    const {
      title,
      children
    } = this.props

    return (
      <MuiThemeProvider muiTheme={getMuiTheme(darkBaseTheme)}>
      <div>
        <style jsx global>{`
					* {
						margin: 0;
						padding: 0;
						text-rendering: geometricPrecision;
						box-sizing: border-box;
					}
					body, html {
        		background: #000;
        		color: #ccc;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
            padding-bottom: 6rem;
      		}
        `}</style>
        <Head>
          <meta charSet='utf-8' />
          <meta name='viewport' content='initial-scale=1.0, width=device-width' />
        </Head>
        <header>
          <Drawer
            docked={false}
            onRequestChange={(open) => this.setState({drawerOpen: open})}
            open={drawerOpen}>
            <MenuItem href='/' leftIcon={<Home/>}>Home</MenuItem>
            <MenuItem href='/admin/jobs' leftIcon={<Event/>}>Jobs</MenuItem>
            <MenuItem href='/admin/clients' leftIcon={<DeviceHub/>}>Clients</MenuItem>
          </Drawer>
          <AppBar
            title={title}
            iconClassNameRight="muidocs-icon-navigation-expand-more"
            onLeftIconButtonTouchTap={() => (this.toggleDrawer())}
          />
        </header>
        <div className='content'>
          { children }
        </div>
        <footer>
          <div className='footer-content'>
          OONI Proteus v0.0.0
          </div>
        </footer>
        <style jsx>{`
          header {
            width: 100%;
            margin-bottom: 20px;
            padding: 0;
          }
          .content {
            font-size: 14px;
            color: #eee;
          }
          footer {
            display: flex;
            flex-wrap: wrap;
            align-items: center;
            margin-top: 20px;
            padding-top: 20px;
            padding-bottom: 20px;
          }
          .footer-content {
            padding: 48px 32px;
            width: 100%;
          }
        `}
        </style>
      </div>
      </MuiThemeProvider>
    )
  }
}
