import React from 'react'
import Head from 'next/head'

import Select from 'react-select'

import Immutable from 'immutable'

import RaisedButton from 'material-ui/RaisedButton'
import FlatButton from 'material-ui/FlatButton'
import Checkbox from 'material-ui/Checkbox'
import DatePicker from 'material-ui/DatePicker'
import TimePicker from 'material-ui/TimePicker'
import Slider from 'material-ui/Slider'
import TextField from 'material-ui/TextField'
import { Card, CardActions, CardHeader, CardTitle, CardText } from 'material-ui/Card'
import { List, ListItem } from 'material-ui/List'
import FloatingActionButton from 'material-ui/FloatingActionButton'
import ContentAdd from 'material-ui/svg-icons/content/add'

import Chip from 'material-ui/Chip'
import Avatar from 'material-ui/Avatar'

import moment from 'moment'

import Layout from '../../../components/layout'
import Session from '../../../components/session'

import { Flex, Box, Grid } from 'reflexbox'

class JobCard extends React.Component {
  static propTypes = {
    onDelete: React.PropTypes.func,
    comment: React.PropTypes.string,
    creationTime: React.PropTypes.string,
    delay: React.PropTypes.number,
    id: React.PropTypes.string,
    state: React.PropTypes.string,
    schedule: React.PropTypes.string,
    target: React.PropTypes.object,
    task: React.PropTypes.object
  }

  render () {
    const {
      state,
      comment,
      creationTime,
      delay,
      id,
      schedule,
      target,
      task,
      onDelete
    } = this.props
    let targetCountries = 'ANY',
        targetPlatforms = 'ANY'
    if (target.countries.length > 0) {
      targetCountries = target.countries.join(',')
    }
    if (target.platforms.length > 0) {
      targetPlatforms = target.platforms.join(',')
    }
    let subtitle = task.test_name
    if (state === 'deleted') {
      subtitle = `[DELETED] ${subtitle}`
    }
    return (
      <Card style={{marginBottom: '20px'}}>
        <CardHeader
          title={comment}
          subtitle={subtitle}
          actAsExpander={true}
          showExpandableButton={true} />
        <CardActions>
          {state !== 'deleted' && <FlatButton label="Delete" onTouchTap={() => {onDelete(id)}}/>}
           <FlatButton label="Edit" onTouchTap={() => {alert('I do nothing')}}/>
        </CardActions>
        <CardText expandable={true}>
          <List>
            <ListItem
                primaryText={schedule}
                secondaryText="Schedule"/>

            <ListItem
                primaryText={''+delay}
                secondaryText="Delay"/>

            <ListItem
                primaryText={creationTime}
                secondaryText="Creation time"/>


            <ListItem
                primaryText={task.test_name}
                secondaryText="Test name"/>
            <ListItem
                primaryText={JSON.stringify(task.arguments)}
                secondaryText="Test arguments"/>

            <ListItem
                primaryText={targetCountries}
                secondaryText="Target countries"/>
            <ListItem
                primaryText={targetPlatforms}
                secondaryText="Target platforms"/>
          </List>

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
    this.onDelete = this.onDelete.bind(this)
  }

  onDelete (jobId) {
    let req = this.state.session.createRequest({baseURL: process.env.EVENTS_URL})
    req.delete(`/api/v1/admin/job/${jobId}`)
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
    let req = this.state.session.createRequest({baseURL: process.env.EVENTS_URL})
    req.get('/api/v1/admin/jobs')
      .then((res) => {
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
      <Layout title="Jobs">
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
                  onDelete={this.onDelete}
                  comment={job.comment}
                  creationTime={job.creation_time}
                  delay={job.delay}
                  id={job.id}
                  schedule={job.schedule}
                  state={job.state}
                  target={job.target}
                  task={job.task} />
                </Grid>
              )
            })}
            <div className='actions'>
              <FloatingActionButton href='/admin/jobs/add'>
                <ContentAdd />
              </FloatingActionButton>
            </div>
          </div>
          <style jsx>{`
          .container {
            max-width: 1024px;
            padding-left: 20px;
            padding-right: 20px;
            margin: auto;
          }
          .actions {
            float: right;
          }
          `}</style>
        </div>
      </Layout>
    )
  }
}
