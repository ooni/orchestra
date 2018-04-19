import axios from 'axios'
import moment from 'moment'

import React from 'react'
import Head from 'next/head'
import Router from 'next/router'

import { Grid } from 'ooni-components'

import Button from 'react-toolbox/lib/button/Button'
import Card from 'react-toolbox/lib/card/Card'
import CardTitle from 'react-toolbox/lib/card/CardTitle'
import CardText from 'react-toolbox/lib/card/CardText'

import List from 'react-toolbox/lib/list/List'
import ListItem from 'react-toolbox/lib/list/ListItem'

import Session from '../../components/session'
import Layout from '../../components/layout'

class ActiveClient extends React.Component {
  static propTypes = {
    clientId: React.PropTypes.string,
    probeAsn: React.PropTypes.string,
    probeCc: React.PropTypes.string,
    platform: React.PropTypes.string,
    softwareName: React.PropTypes.string,
    softwareVersion: React.PropTypes.string,
    supportedTests: React.PropTypes.string,
    networkType: React.PropTypes.string,
    language: React.PropTypes.string,
    availableBandwidth: React.PropTypes.number,
    token: React.PropTypes.string,
    lastUpdated: React.PropTypes.string,
    created: React.PropTypes.object,
  }
  constructor (props) {
    super(props)
    this.state = {
      isOpen: false
    }
  }

  render () {
    const {
      clientId,
      probeAsn,
      probeCc,
      platform,
      softwareName,
      softwareVersion,
      supportedTests,
      networkType,
      language,
      availableBandwidth,
      token,
      lastUpdated,
      created
    } = this.props
    const {
      isOpen
    } = this.state

    const title = `${clientId.slice(-5)} (${probeAsn}, ${probeCc})`
    const subtitle = `last updated ${moment(lastUpdated).fromNow()}`
    return (
      <Card style={{position: 'relative'}}>
        <div style={{position: 'absolute', right: 0}} onClick={() => {this.setState({isOpen: !this.state.isOpen})}}>
          {isOpen && <Button icon='keyboard_arrow_up' />}
          {!isOpen && <Button icon='keyboard_arrow_down' />}
        </div>
        <CardTitle
          title={title}
          subtitle={subtitle} />
        <CardText>
          {isOpen && <List>
            <ListItem
                caption={probeAsn}
                legend="ASN"/>

            <ListItem
                caption={probeCc}
                legend="Country"/>

            <ListItem
                caption={platform}
                legend="Platform"/>

            <ListItem
                caption={softwareName + ' v' + softwareVersion}
                legend="Software version"/>

            <ListItem
                caption={supportedTests}
                legend="Supported tests"/>
            <ListItem
                caption={networkType}
                legend="Network Type"/>

            <ListItem
                caption={language}
                legend="Language"/>

            <ListItem
                caption={''+availableBandwidth}
                legend="Bandwidth"/>
            <ListItem
                caption={token}
                legend="Token"/>
            <ListItem
                caption={moment(lastUpdated).fromNow()}
                legend="Updated"/>
            <ListItem
                caption={moment(created).fromNow()}
                legend="Created"/>
          </List>}

        </CardText>
      </Card>
    )
  }
}

const MetadataRow = ({metadata}) => {
  return (
    <div style={{paddingBottom: '20px'}}>
      <div>Total Client Count: {metadata.total_client_count}</div>
      <div>Countries: {metadata.client_countries.map((d) => {
        return <span style={{paddingRight: '30px'}}>{d.probe_cc} ({d.count})</span>
      })}
      </div>
  </div>
  )
}

const Pagination = ({limit, offset, nextPage, prevPage}) => {
  return <div style={{paddingTop: '30px'}}>
    {offset > 0 && <a onClick={prevPage}>← previous page</a>}
    <a style={{paddingLeft: '20px'}} onClick={nextPage}>next page →</a>
  </div>
}
export default class AdminClients extends React.Component {

  constructor (props) {
    super(props)
    this.state = {
      session: new Session(),
      metadata: {},
      results: [],
      limit: 100,
      offset: 0
    }
    this.getNextPage = this.getNextPage.bind(this)
    this.getPrevPage = this.getPrevPage.bind(this)
  }

  componentDidMount() {
    if (this.state.session.redirectIfInvalid() === true) {
      return
    }
    let req = this.state.session.createRequest({baseURL: process.env.REGISTRY_URL})
    req.get('/api/v1/admin/clients')
      .then((res) => {
        const { metadata, results } = res.data
        this.setState({metadata, results})
      })
  }

  getPrevPage() {
    let req = this.state.session.createRequest({baseURL: process.env.REGISTRY_URL})
    let newOffset = this.state.offset - this.state.limit
    if (newOffset < 0) {
      newOffset = 0
    }
    const params = {offset: newOffset, limit: this.state.limit}
    req.get('/api/v1/admin/clients', {params})
      .then((res) => {
        const { metadata, results } = res.data
        this.setState({metadata, results, offset: newOffset})
      })
  }

  getNextPage() {
    let req = this.state.session.createRequest({baseURL: process.env.REGISTRY_URL})
    const newOffset = this.state.offset + this.state.limit
    const params = {offset: newOffset, limit: this.state.limit}
    req.get('/api/v1/admin/clients', {params})
      .then((res) => {
        const { metadata, results } = res.data
        this.setState({metadata, results, offset: newOffset})
      })
  }

  render () {
    const {
      results,
      metadata,
      limit,
      offset
    } = this.state

    return (
      <Layout title="Active clients">
        <Head>
          <title>Active Clients - OONI Proteus</title>
        </Head>
        <style jsx>{`
        .container {
            max-width: 1024px;
            padding-left: 20px;
            padding-right: 20px;
            margin: auto;
        }
        h1 {
          margin-bottom: 20px;
        }
        `}</style>
        <div className='container'>
          {metadata.count && <MetadataRow metadata={metadata} />}
          {results && results.map((d) => {
            return (
              <Grid col={4} px={2}>
              <ActiveClient
                clientId={d.client_id}
                probeAsn={d.probe_asn}
                probeCc={d.probe_cc}
                platform={d.platform}
                softwareName={d.software_name}
                softwareVersion={d.software_version}
                supportedTests={d.supported_tests}
                networkType={d.network_type}
                language={d.language}
                availableBandwidth={d.available_bandwidth}
                token={d.token}
                lastUpdated={d.last_updated}
                created={d.creation_time} />
            </Grid>
            )
          })}
          <Pagination
            limit={limit}
            offset={offset}
            nextPage={this.getNextPage}
            prevPage={this.getPrevPage}
            />
        </div>
      </Layout>
    )
  }
}
