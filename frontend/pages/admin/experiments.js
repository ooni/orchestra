
import React from 'react'
import PropTypes from 'prop-types'

import Head from 'next/head'
import Link from 'next/link'

import Immutable from 'immutable'

import Avatar from 'material-ui/Avatar'

import Button from 'material-ui/Button'

import moment from 'moment'

import {
  Flex,
  Box,
  Container,
  Heading,
  Text
} from 'ooni-components'

import Layout from '../../components/layout'
import Session from '../../components/session'

import JobCard from '../../components/ui/jobs/JobCard'

class AdminExperimentsIndex extends React.Component {
  constructor (props) {
    super(props)
    this.state = {
      jobList: [],
      error: null,
      session: new Session(),
    }
    this.onDelete = this.onDelete.bind(this)
  }

  onDelete (jobId) {
    let req = this.state.session.createRequest({baseURL: process.env.ORCHESTRATE_URL})
    req.delete(`/api/v1/admin/experiment/${jobId}`)
      .then((res) => {
        const newJobList = this.state.jobList.filter((job) => (job.id !== jobId))
        this.setState({
          jobList: newJobList
        })
      })
      .catch((err) => {
        // XXX handle errors
      })
  }

  componentDidMount() {
    if (this.state.session.redirectIfInvalid()) {
      return
    }
    let req = this.state.session.createRequest({baseURL: process.env.ORCHESTRATE_URL})
    req.get('/api/v1/admin/experiments')
      .then((res) => {
        this.setState({
          jobList: res.data.jobs || [],
          error: null
        })
    }).catch((err) => {
        this.setState({
          error: err
        })
    })
  }

  render () {
    const {
      jobList
    } = this.state

    return (
      <Layout>
        <Head>
          <title>Experiments - OONI Proteus</title>
        </Head>

          <Container style={{position: 'relative'}}>
            <Heading h={1}>List of experiments</Heading>
            {jobList.length == 0 && <Heading h={5}>No experiment has been scheduled</Heading>}
            <Link href='/admin/experiment/new'><Button color='primary' variant='raised'>Create new experiment</Button></Link>
            {jobList.map((job) => {
              return (
                <JobCard
                  key={job.id}
                  onDelete={this.onDelete}
                  comment={job.comment}
                  creationTime={job.creation_time}
                  delay={job.delay}
                  id={job.id}
                  schedule={job.schedule}
                  state={job.state}
                  target={job.target}
                  alertData={job.alert}
                  task={job.task} />
              )
            })}

          </Container>
      </Layout>
    )
  }
}

export default AdminExperimentsIndex
