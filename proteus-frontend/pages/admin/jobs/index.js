import React from 'react'
import Head from 'next/head'
import Link from 'next/link'
import Select from 'react-select'

import Immutable from 'immutable'

import Avatar from 'react-toolbox/lib/avatar/Avatar'
import Button from 'react-toolbox/lib/button/Button'
import Card from 'react-toolbox/lib/card/Card'
import CardActions from 'react-toolbox/lib/card/CardActions'
import CardTitle from 'react-toolbox/lib/card/CardTitle'
import CardText from 'react-toolbox/lib/card/CardText'

import List from 'react-toolbox/lib/list/List'
import ListItem from 'react-toolbox/lib/list/ListItem'

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

  constructor (props) {
    super(props)
    this.state = {
      isOpen: false
    }
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
      alertData,
      onDelete
    } = this.props
    const {
      isOpen
    } = this.state

    let targetCountries = 'ANY',
        targetPlatforms = 'ANY'
    if (target.countries.length > 0) {
      targetCountries = target.countries.join(',')
    }
    if (target.platforms.length > 0) {
      targetPlatforms = target.platforms.join(',')
    }
    let subtitle
    if (task) {
      subtitle = task.test_name
    }
    if (state === 'deleted') {
      subtitle = `[DELETED] ${subtitle}`
    }

    let cardAvatarValue, cardAvatar
    if (task) {
      cardAvatarValue = 'assignment'
    } else {
      cardAvatarValue = 'message'
    }
    cardAvatar = <Avatar icon={cardAvatarValue} style={{paddingTop: '8px'}}/>
    return (
      <Card style={{position: 'relative'}}>
        <div style={{position: 'absolute', right: 0}} onClick={() => {this.setState({isOpen: !this.state.isOpen})}}>
          {isOpen && <Button icon='keyboard_arrow_up' />}
          {!isOpen && <Button icon='keyboard_arrow_down' />}
        </div>
        <CardTitle
          title={comment}
          avatar={cardAvatar}
          subtitle={subtitle}
          />
        <CardActions>
          {state !== 'deleted' && <Button label="Delete" onClick={() => {onDelete(id)}}/>}
           <Button label="Edit" onClick={() => {alert('I do nothing')}}/>
        </CardActions>
        <CardText>
          {isOpen && <List>
            {alertData && <ListItem
                caption={alertData.message}
                legend="Message"/>
            }
            {alertData && <ListItem
                caption={JSON.stringify(alertData.extra)}
                legend="Alert Extra"/>
            }

            {task && <ListItem
                caption={task.test_name}
                legend="Test name"/>
            }
            {task && <ListItem
                caption={JSON.stringify(task.arguments)}
                legend="Test arguments"/>
            }

            <ListItem
                caption={schedule}
                legend="Schedule"/>

            <ListItem
                caption={''+delay}
                legend="Delay"/>

            <ListItem
                caption={creationTime}
                legend="Creation time"/>

            <ListItem
                caption={targetCountries}
                legend="Target countries"/>
            <ListItem
                caption={targetPlatforms}
                legend="Target platforms"/>
          </List>}

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
      session: new Session(),
      actionButtonOpen: false
    }
    this.onDelete = this.onDelete.bind(this)
    this.toggleAction = this.toggleAction.bind(this)
  }

  toggleAction () {
    this.setState({actionButtonOpen: !this.state.actionButtonOpen})
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
      jobList,
      actionButtonOpen
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
                  alertData={job.alert}
                  task={job.task} />
                </Grid>
              )
            })}
            <div className='actions'>
              <Flex column>
                {actionButtonOpen && <Box pt={2}>
                  <Link href='/admin/jobs/add_alert'><Button floating icon='message' mini /></Link>
                </Box>}
                {actionButtonOpen && <Box pt={2}>
                  <Link href='/admin/jobs/add_task'><Button floating icon='assignment' mini /></Link>
                </Box>}

                <Box pt={2} onClick={() => this.toggleAction()}>
                  {!actionButtonOpen && <Button floating icon='add' accent />}
                  {actionButtonOpen && <Button floating icon='clear' accent />}
                </Box>
              </Flex>
            </div>
          </div>
          <style jsx>{`
          .container {
            max-width: 1024px;
            padding-left: 20px;
            padding-right: 20px;
            margin: auto;
            position: relative;
            min-height: 50vh;
          }
          .actions {
            position: absolute;
            bottom: 0;
            right: 0;
          }
          `}</style>
        </div>
      </Layout>
    )
  }
}
