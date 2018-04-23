import React from 'react'
import PropTypes from 'prop-types'

import Router from 'next/router'
import NProgress from 'nprogress'

import Toolbar from 'material-ui/Toolbar'
import Button from 'material-ui/Button'


import Link from 'next/link'
import Head from 'next/head'

import {
  theme,
  Provider,
  Container
} from 'ooni-components'

import styled from 'styled-components'

import pkgJson from '../package.json'

Router.onRouteChangeStart = (url) => NProgress.start()
Router.onRouteChangeComplete = () => NProgress.done()
Router.onRouteChangeError = () => NProgress.done()

const Footer = styled.div`
  position: fixed;
  left: 0;
  bottom: 0;
  width: 100%;
  padding-top: 20px;
  padding-bottom: 20px;
  text-align: center;
  background-color: ${props => props.theme.colors.gray8};
  color: ${props => props.theme.colors.white};
`

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
      <Provider theme={theme}>
        <div>
          <Head>
            <meta charSet='utf-8' />
            <meta name='viewport' content='initial-scale=1.0, width=device-width' />
            <link rel='stylesheet' href='/static/vendor/nprogress.css'/>
          </Head>
          <header>
            <Toolbar>
            <Link href='/'><Button>Home</Button></Link>
            <Link href='/admin/jobs'><Button>Alerts</Button></Link>
            <Link href='/admin/experiment/new'><Button>Experiments</Button></Link>
            <Link href='/admin/clients'><Button>Clients</Button></Link>
            </Toolbar>
          </header>
          <div className='content'>
            { children }
          </div>
          <Footer>
            <Container>
            OONI Orchestra {pkgJson['version']}
            </Container>
          </Footer>
        </div>
      </Provider>
    )
  }
}
