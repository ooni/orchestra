import React from 'react'
import Head from 'next/head'

import Layout from '../components/layout'

export default () => {
  return (
    <Layout>
      <Head>
        <title>OPOS</title>
      </Head>
      <div className='hero'>
        <div className='hero-unit'>
          <img height='100px' src='/static/proteus-white.png' />
          <h1>Proteus</h1>
        </div>
      </div>
      <style jsx>{`
        .hero {
          display: flex;
          margin: 0 auto;
          align-items: center;
          min-height: calc(100vh - 95px);
          justify-content: center;
          position: relative;
        }
        h1 {
          font-size: 80px;
          font-weight: 200;
          line-height: 80px;
        }
        .hero-unit {
          text-align: center;
        }
      `}</style>
    </Layout>
  )
}
