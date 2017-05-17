import React from 'react'
import Head from 'next/head'
import Router from 'next/router'

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
import Divider from 'material-ui/Divider'

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
        <CardHeader
          title={comment}
          subtitle={task.test_name}
          actAsExpander={true}
          showExpandableButton={true} />
        <CardActions>
           <FlatButton label="Delete" onTouchTap={() => {alert('I do nothing')}}/>
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
                primaryText={target.countries.join(',')}
                secondaryText="Target countries"/>
            <ListItem
                primaryText={target.platforms.join(',')}
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
            <h1 style={{marginBottom: 20}}>Currently scheduled jobs</h1>
            {jobList.map((job) => {
              return (
                <Grid col={6} px={2}>
                <JobCard
                  key={job.id}
                  comment={job.comment}
                  creationTime={job.creation_time}
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
