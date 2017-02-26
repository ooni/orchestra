import axios from 'axios'

import React from 'react'
import Head from 'next/head'

import Layout from '../../components/layout'

// XXX protect this with some auth
export default class AdminClients extends React.Component {

  static async getInitialProps () {
    let req = axios.create({baseURL: REGISTRY_URL})
    const res = await req.get('/api/v1/clients')
    return { clients: res.data }
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
          <h1>Active clients</h1>
          <table>
          <thead>
            <tr>
              <td>Client ID</td>
              <td>Probe ASN</td>
              <td>Probe CC</td>
              <td>Platform</td>
              <td>Software</td>
              <td>Supported Tests</td>
              <td>Network type</td>
              <td>Available Bandwidth</td>
              <td>Token</td>
              <td>Probe ID</td>
              <td>Probe Family</td>
              <td>Last Updated</td>
              <td>Creation Time</td>
            </tr>
          </thead>
          <tbody>
          {this.props.clients['active_clients'].map((d) => {
            return (
              <tr>
              <td>{d.client_id}</td>
              <td>{d.probe_asn}</td>
              <td>{d.probe_cc}</td>
              <td>{d.platform}</td>
              <td>{d.software_name} - {d.software_version}</td>
              <td>{d.supported_tests}</td>
              <td>{d.network_type}</td>
              <td>{d.available_bandwidth}</td>
              <td>{d.token}</td>
              <td>{d.probe_id}</td>
              <td>{d.probe_family}</td>
              <td>{d.last_updated}</td>
              <td>{d.creation_time}</td>
              </tr>
            )
          })}
          </tbody>
          </table>
        </div>
      </Layout>
    )
  }
}
