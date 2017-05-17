import React from 'react'
import Head from 'next/head'
import Router from 'next/router'

import Select from 'react-select'

import Immutable from 'immutable'

import RaisedButton from 'material-ui/RaisedButton'
import Checkbox from 'material-ui/Checkbox'
import DatePicker from 'material-ui/DatePicker'
import TimePicker from 'material-ui/TimePicker'
import Slider from 'material-ui/Slider'
import TextField from 'material-ui/TextField'
import {Card, CardActions, CardTitle, CardText} from 'material-ui/Card'
import Chip from 'material-ui/Chip'
import Avatar from 'material-ui/Avatar'

import moment from 'moment'

import Layout from '../../../components/layout'
import Session from '../../../components/session'

import { Flex, Box, Grid } from 'reflexbox'

class JobCard extends React.Component {
  static propTypes = {
    comment: React.PropTypes.string,
    creationTime: React.PropTypes.string,
    delay: React.PropTypes.number,
    id: React.PropTypes.string,
    schedule: React.PropTypes.string,
    target: React.PropTypes.object,
    task: React.PropTypes.object
  }
  render () {
    const {
      comment,
      creationTime,
      delay,
      id,
      schedule,
      target,
      task
    } = this.props
    return (
      <Card style={{marginBottom: '20px'}}>
        <CardTitle title={comment}/>
        <CardText>
          <p><strong>Creation Time:</strong> {creationTime}</p>

          <p><strong>Delay:</strong> {delay}</p>

          {/*<p><strong>Id:</strong> {id}</p> */}

          <p><strong>Schedule:</strong> {schedule}</p>

          <div style={{display: 'flex', flexWrap: 'wrap'}}><strong>Target Countries:</strong> 
          {target.countries.map((country) => {
            return (
              <Chip style={{margin: 4}}>
                <Avatar>{country}</Avatar>
                country name
              </Chip>
            )
          })}
          </div>

          <p>
          <strong>Test Name: </strong>
          {task.test_name}
          </p>
          <p>
          <strong>Arguments: </strong>
          {JSON.stringify(task.arguments)}
          </p>
        </CardText>
      </Card>
    )
  }
}

export default class AdminJobsIndex extends React.Component {

  constructor (props) {
    super(props)
    this.state = {
      jobList: [],
      error: null,
      session: new Session()
    }
  }

  componentDidMount() {
    if (this.state.session.isValid() === false) {
      Router.push('/admin/login?from="'+Router.route+'"')
      return
    }
    let req = this.state.session.createRequest({baseURL: process.env.EVENTS_URL})
    req.get('/api/v1/admin/jobs')
      .then((res) => {
        console.log(res.data)
        this.setState({
          jobList: res.data.jobs,
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
          <title>Jobs - OONI Proteus</title>
        </Head>

        <div>
          <div className='container'>
            {jobList.map((job) => {
              return (
                <Grid col={6} px={2}>
                <JobCard
                  key={job.id}
                  comment={job.comment}
                  creationTime={job.creationTime}
                  delay={job.delay}
                  id={job.id}
                  schedule={job.schedule}
                  target={job.target}
                  task={job.task} />
                </Grid>
              )
            })}
          </div>
          <style jsx>{`
          .container {
            max-width: 1024px;
            padding-left: 20px;
            padding-right: 20px;
            margin: auto;
          }
          `}</style>
        </div>
      </Layout>
    )
  }
}
