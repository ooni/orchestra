import Head from 'next/head'
import NProgress from 'nprogress'
import Router from 'next/router'

Router.onRouteChangeStart = () => NProgress.start()
Router.onRouteChangeComplete = () => NProgress.done()
Router.onRouteChangeError = () => NProgress.done()

export default () => (
  <div>
		<Head>
			<meta charSet='utf-8' />
			<meta name='viewport' content='initial-scale=1.0, width=device-width' />
			<link href='/static/theme.css' rel='stylesheet' />
			<link href='/static/vendor/material-icons/material-design-icons.css' rel='stylesheet' />
		</Head>
		<style jsx global>{`
			* {
				margin: 0;
				padding: 0;
				text-rendering: geometricPrecision;
				box-sizing: border-box;
			}
			body, html {
				background-color: white;
				color: #1c1c1c;
				font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
				padding-bottom: 6rem;
			}
      /* loading progress bar styles */
      #nprogress {
        pointer-events: none;
      }
      #nprogress .bar {
        background: #0588CB;
        position: fixed;
        z-index: 1031;
        top: 0;
        left: 0;
        width: 100%;
        height: 2px;
      }
      #nprogress .peg {
        display: block;
        position: absolute;
        right: 0px;
        width: 100px;
        height: 100%;
        box-shadow: 0 0 10px #0588CB, 0 0 5px #0588CB;
        opacity: 1.0;
        transform: rotate(3deg) translate(0px, -4px);
      }
    `}</style>
  </div>
)
