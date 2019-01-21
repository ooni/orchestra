import React from 'react'
import PropTypes from 'prop-types'

import Head from 'next/head'
import Router from 'next/router'

import {
  Flex,
  Box,
  Grid,
  Container
} from 'ooni-components'

import Button from '@material-ui/core/Button'

import Card from '@material-ui/core/Card'
import CardHeader from '@material-ui/core/CardHeader'
import CardContent from '@material-ui/core/CardContent'
import CardActions from '@material-ui/core/CardActions'

import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'

import KeyboardArrowUpIcon from '@material-ui/icons/KeyboardArrowUp'
import KeyboardArrowDownIcon from '@material-ui/icons/KeyboardArrowDown'

import Session from '../../components/session'
import Layout from '../../components/layout'

import axios from 'axios'
import moment from 'moment'

class ActiveClient extends React.Component {
  static propTypes = {
    clientId: PropTypes.string,
    probeAsn: PropTypes.string,
    probeCc: PropTypes.string,
    platform: PropTypes.string,
    softwareName: PropTypes.string,
    softwareVersion: PropTypes.string,
    supportedTests: PropTypes.string,
    networkType: PropTypes.string,
    language: PropTypes.string,
    availableBandwidth: PropTypes.number,
    token: PropTypes.string,
    lastUpdated: PropTypes.string,
    created: PropTypes.object,
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
          {isOpen && <Button><KeyboardArrowUpIcon/></Button>}
          {!isOpen && <Button><KeyboardArrowDownIcon/></Button>}
        </div>
        <CardHeader
          title={title}
          subheader={subtitle} />
        <CardContent>
          {isOpen && <List>
            <ListItem>
            <ListItemText
                primary={probeAsn}
                secondary="ASN"/>
            </ListItem>

            <ListItem>
            <ListItemText
                primary={probeCc}
                secondary="Country"/>
            </ListItem>

            <ListItem>
            <ListItemText
                primary={platform}
                secondary="Platform"/>
            </ListItem>

            <ListItem>
            <ListItemText
                primary={softwareName + ' v' + softwareVersion}
                secondary="Software version"/>
            </ListItem>

            <ListItem>
            <ListItemText
                primary={supportedTests}
                secondary="Supported tests"/>
            </ListItem>
            <ListItem>
            <ListItemText
                primary={networkType}
                secondary="Network Type"/>
            </ListItem>

            <ListItem>
            <ListItemText
                primary={language}
                secondary="Language"/>
            </ListItem>

            <ListItem>
            <ListItemText
                primary={''+availableBandwidth}
                secondary="Bandwidth"/>
            </ListItem>
            <ListItem>
            <ListItemText
                primary={token}
                secondary="Token"/>
            </ListItem>
            <ListItem>
            <ListItemText
                primary={moment(lastUpdated).fromNow()}
                secondary="Updated"/>
            </ListItem>
            <ListItem>
            <ListItemText
                primary={moment(created).fromNow()}
                secondary="Created"/>
            </ListItem>
          </List>}

        </CardContent>
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
        <Container>
          {metadata.count && <MetadataRow metadata={metadata} />}
          <Flex wrap>
          {results && results.map((d) => {
            return (
              <Box key={d.client_id} width={1/3}>
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
            </Box>
            )
          })}
          </Flex>
          <Pagination
            limit={limit}
            offset={offset}
            nextPage={this.getNextPage}
            prevPage={this.getPrevPage}
            />
        </Container>
      </Layout>
    )
  }
}
