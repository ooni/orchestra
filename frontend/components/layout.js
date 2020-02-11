import React from 'react'
import PropTypes from 'prop-types'

import ThemeProvider from 'react-toolbox/lib/ThemeProvider'
import theme from '../public/static/theme'

import Drawer from 'react-toolbox/lib/drawer/Drawer'
import MenuItem from 'react-toolbox/lib/menu/MenuItem'
import AppBar from 'react-toolbox/lib/app_bar/AppBar'

import Link from 'next/link'

import Meta from './meta'

import pkgJson from '../package.json'

export default class extends React.Component {

  static propTypes = {
    children: PropTypes.node.isRequired,
    title: PropTypes.string
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
    let {
      title,
      children
    } = this.props
    if (!title) {
      title = "Proteus"
    }
    return (
      <ThemeProvider theme={theme}>
      <div>
        <Meta />
        <header>
          <Drawer
            onOverlayClick={() => this.setState({drawerOpen: false})}
            active={drawerOpen}>
            <Link href='/'><MenuItem icon='home'>Home</MenuItem></Link>
            <Link href='/admin/jobs'><MenuItem icon='event'>Jobs</MenuItem></Link>
            <Link href='/admin/clients'><MenuItem icon='device_hub'>Clients</MenuItem></Link>
          </Drawer>
          <AppBar
            title={title}
            leftIcon='menu'
            onLeftIconClick={() => (this.toggleDrawer())}
          />
        </header>
        <div className='content'>
          { children }
        </div>
        <footer>
          <div className='footer-content'>
          OONI Proteus {pkgJson['version']}
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
            color: #1c1c1c;
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
      </ThemeProvider>
    )
  }
}
