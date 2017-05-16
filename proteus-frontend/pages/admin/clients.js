import axios from 'axios'
import moment from 'moment'

import React from 'react'
import Head from 'next/head'
import Router from 'next/router'

import {
  Table,
  TableHeader,
	TableHeaderColumn,
  TableBody,
  TableRow,
  TableRowColumn,
  TableCell
} from 'material-ui/Table'

import Session from '../../components/session'
import Layout from '../../components/layout'

export default class AdminClients extends React.Component {

  constructor (props) {
    super(props)
    this.state = {
      session: new Session(),
      clients: {}
    }
  }

  componentDidMount() {
    if (this.state.session.isValid() === false) {
      Router.push('/admin/login')
      return
    }
    let req = this.state.session.createRequest({baseURL: process.env.REGISTRY_URL})
    req.get('/api/v1/admin/clients')
      .then((res) => {
        console.log(res.data)
        this.setState({clients: res.data})
      })
  }

  render () {
    return (
      <Layout>
        <Head>
          <title>Active Clients - OONI Proteus</title>
        </Head>
        <style jsx>{`
        .active-clients {
          padding: 20px;
        }
        `}</style>
        <div className="active-clients">
					<Table
            multiSelectable={true}>
          <TableHeader>
						<TableRow>
             	<TableHeaderColumn colSpan='11' style={{textAlign: 'center'}}>
                Active clients
              </TableHeaderColumn>
            </TableRow>
            <TableRow>
              <TableHeaderColumn tooltip="Client ID">Client ID</TableHeaderColumn>
              <TableHeaderColumn tooltip="Probe ASN">ASN</TableHeaderColumn>
              <TableHeaderColumn tooltip="Probe Country Code">CC</TableHeaderColumn>
              <TableHeaderColumn tooltip="Probe Platform">Platform</TableHeaderColumn>
              <TableHeaderColumn tooltip="Software Version">Software</TableHeaderColumn>
              <TableHeaderColumn tooltip="Supported Tests">Tests</TableHeaderColumn>
              <TableHeaderColumn>Network type</TableHeaderColumn>
              <TableHeaderColumn tooltip="Available Bandwidth">Bandwidth</TableHeaderColumn>
              <TableHeaderColumn>Token</TableHeaderColumn>
              <TableHeaderColumn>Last Updated</TableHeaderColumn>
              <TableHeaderColumn>Creation Time</TableHeaderColumn>
            </TableRow>
          </TableHeader>
					<TableBody>
          {this.state.clients['active_clients'] && this.state.clients['active_clients'].map((d) => {
            return (
							<TableRow key={d.client_id}>
								<TableRowColumn>{d.client_id}</TableRowColumn>
								<TableRowColumn>{d.probe_asn}</TableRowColumn>
								<TableRowColumn>{d.probe_cc}</TableRowColumn>
								<TableRowColumn>{d.platform}</TableRowColumn>
								<TableRowColumn>{d.software_name} - {d.software_version}</TableRowColumn>
								<TableRowColumn>{d.supported_tests}</TableRowColumn>
								<TableRowColumn>{d.network_type}</TableRowColumn>
								<TableRowColumn>{d.available_bandwidth}</TableRowColumn>
								<TableRowColumn>{d.token}</TableRowColumn>
								<TableRowColumn>
                {moment(d.last_updated).fromNow()}
                </TableRowColumn>
								<TableRowColumn>
                {moment(d.creation_time).fromNow()}
                </TableRowColumn>
              </TableRow>
            )
          })}
					</TableBody>
          </Table>
        </div>
      </Layout>
    )
  }
}
