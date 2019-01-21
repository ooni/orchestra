import React from 'react'
import Head from 'next/head'

import Layout from '../components/layout'

import { Text, Heading, Container } from 'ooni-components'

export default () => {
  return (
    <Layout>
      <Head>
        <title>OONI Orchestra</title>
      </Head>
      <Container>
        <Heading h={1}>Welcome to the OONI Orchestra!</Heading>
        <Text>registry: {process.env.REGISTRY_URL}</Text>
        <Text>orchestrate: {process.env.ORCHESTRATE_URL}</Text>
      </Container>
    </Layout>
  )
}
