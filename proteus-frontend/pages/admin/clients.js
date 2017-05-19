import axios from 'axios'
import moment from 'moment'

import React from 'react'
import Head from 'next/head'
import Router from 'next/router'

import RaisedButton from 'material-ui/RaisedButton'
import FlatButton from 'material-ui/FlatButton'
import Checkbox from 'material-ui/Checkbox'
import DatePicker from 'material-ui/DatePicker'
import TimePicker from 'material-ui/TimePicker'
import Slider from 'material-ui/Slider'
import TextField from 'material-ui/TextField'
import { Card, CardActions, CardHeader, CardTitle, CardText } from 'material-ui/Card'
import { List, ListItem } from 'material-ui/List'
import Divider from 'material-ui/Divider'
import { Flex, Box, Grid } from 'reflexbox'

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
    availableBandwidth: React.PropTypes.number,
    token: React.PropTypes.string,
    lastUpdated: React.PropTypes.string,
    created: React.PropTypes.object,
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
      availableBandwidth,
      token,
      lastUpdated,
      created
    } = this.props
    const title = `${clientId.slice(-5)} (${probeAsn}, ${probeCc})`
    const subtitle = `last updated ${moment(lastUpdated).fromNow()}`
    return (
      <Card style={{marginBottom: '20px'}}>
        <CardHeader
          title={title}
          subtitle={subtitle}
          actAsExpander={true}
          showExpandableButton={true} />
        <CardText expandable={true}>
          <List>
            <ListItem
                primaryText={probeAsn}
                secondaryText="ASN"/>

            <ListItem
                primaryText={probeCc}
                secondaryText="Country"/>

            <ListItem
                primaryText={platform}
                secondaryText="Platform"/>

            <ListItem
                primaryText={softwareName + ' v' + softwareVersion}
                secondaryText="Software version"/>

            <ListItem
                primaryText={supportedTests}
                secondaryText="Supported tests"/>
            <ListItem
                primaryText={networkType}
                secondaryText="Network Type"/>
            <ListItem
                primaryText={''+availableBandwidth}
                secondaryText="Bandwidth"/>
            <ListItem
                primaryText={token}
                secondaryText="Token"/>
            <ListItem
                primaryText={moment(lastUpdated).fromNow()}
                secondaryText="Updated"/>
            <ListItem
                primaryText={moment(created).fromNow()}
                secondaryText="Created"/>
          </List>

        </CardText>
      </Card>
    )
  }
}
export default class AdminClients extends React.Component {

  constructor (props) {
    super(props)
    this.state = {
      session: new Session(),
      clients: {}
    }
  }

  componentDidMount() {
    if (this.state.session.redirectIfInvalid() === true) {
      return
    }
    let req = this.state.session.createRequest({baseURL: process.env.REGISTRY_URL})
    req.get('/api/v1/admin/clients')
      .then((res) => {
        this.setState({clients: res.data})
      })
  }

  render () {
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
          {this.state.clients['active_clients'] && this.state.clients['active_clients'].map((d) => {
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
                availableBandwidth={d.available_bandwidth}
                token={d.token}
                lastUpdated={d.last_updated}
                created={d.creation_time} />
            </Grid>
            )
          })}
        </div>
      </Layout>
    )
  }
}
