import React from 'react'

import darkBaseTheme from 'material-ui/styles/baseThemes/darkBaseTheme'
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider'
import getMuiTheme from 'material-ui/styles/getMuiTheme'

import Link from 'next/link'
import Head from 'next/head'

import './tapEvents'

export default class extends React.Component {

  static propTypes = {
    children: React.PropTypes.array.isRequired
  }

  render () {
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
          <div className='nav-container'>
            <div className='nav'>
              <Link href='/'><a>Home</a></Link>
              <Link href='/admin/clients'><a>Admin/Clients</a></Link>
              <Link href='/admin/jobs'><a>Admin/Jobs</a></Link>
            </div>
          </div>
        </header>
        <div className='content'>
          { this.props.children }
        </div>
        <footer>
          OONI Proteus v0.0.0
        </footer>
        <style jsx>{`
          header {
            max-width: 900px;
            padding: 30px 0;
            margin: auto;
            position: relative;
          }
          header .nav {
            position: absolute;
            padding: 10px;
            padding-right: 0;
            top: 50%;
            left: 0px;
            transform: translateY(-50%);
          }
          .nav > a {
            padding: 10px;
            font-size: 12px;
            text-transform: uppercase;
            font-weight: normal;
            color: #ccc;
            text-decoration: none;
          }
          .nav > a:hover {
            color: #fff;
          }
          .content {
            font-size: 14px;
            color: #eee;
          }
        `}
        </style>
      </div>
      </MuiThemeProvider>
    )
  }
}
