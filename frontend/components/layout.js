import React from 'react'
import PropTypes from 'prop-types'

import Router from 'next/router'
import NProgress from 'nprogress'

import Toolbar from 'material-ui/Toolbar'
import Button from 'material-ui/Button'

import Link from 'next/link'
import Head from 'next/head'

import pkgJson from '../package.json'

Router.onRouteChangeStart = (url) => {
  console.log("Loading ", url)
  NProgress.start()
}

Router.onRouteChangeComplete = () => NProgress.done()
Router.onRouteChangeError = () => NProgress.done()

export default class extends React.Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
    title: PropTypes.string
  }

  render () {
    let {
      title,
      children
    } = this.props
    if (!title) {
      title = "Orchestra"
    }
    return (
      <div>
        <Head>
          <meta charSet='utf-8' />
          <meta name='viewport' content='initial-scale=1.0, width=device-width' />
        </Head>
        <header>
          <Toolbar>
          <Link href='/'><Button>Home</Button></Link>
          <Link href='/admin/jobs'><Button>Jobs</Button></Link>
          <Link href='/admin/clients'><Button>Clients</Button></Link>
          </Toolbar>
        </header>
        <div className='content'>
          { children }
        </div>
        <footer>
          <div className='footer-content'>
          OONI Proteus {pkgJson['version']}
          </div>
        </footer>
      </div>
    )
  }
}
