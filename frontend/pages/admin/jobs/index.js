import React from 'react'
import PropTypes from 'prop-types'

import Head from 'next/head'
import Link from 'next/link'

import Immutable from 'immutable'

import Avatar from 'material-ui/Avatar'

import Button from 'material-ui/Button'
import AddIcon from '@material-ui/icons/Add'
import ClearIcon from '@material-ui/icons/Clear'
import DeleteIcon from '@material-ui/icons/Delete'
import MessageIcon from '@material-ui/icons/Message'
import AssignmentIcon from '@material-ui/icons/Assignment'
import KeyboardArrowUpIcon from '@material-ui/icons/KeyboardArrowUp'
import KeyboardArrowDownIcon from '@material-ui/icons/KeyboardArrowDown'

import Card, { CardHeader, CardContent, CardActions } from 'material-ui/Card'

import List, { ListItem, ListItemText } from 'material-ui/List'

import moment from 'moment'

import { Flex, Box } from 'ooni-components'
import { Container } from 'ooni-components'

import Layout from '../../../components/layout'
import Session from '../../../components/session'

class JobCard extends React.Component {
  static propTypes = {
    onDelete: PropTypes.func,
    comment: PropTypes.string,
    creationTime: PropTypes.string,
    delay: PropTypes.number,
    id: PropTypes.string,
    state: PropTypes.string,
    schedule: PropTypes.string,
    target: PropTypes.object,
    task: PropTypes.object
  }

  constructor (props) {
    console.log("Callin contr")
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

    let cardAvatar
    if (task) {
      cardAvatar = <Avatar style={{paddingTop: '8px'}}><AssignmentIcon /></Avatar>
    } else {
      cardAvatar = <Avatar style={{paddingTop: '8px'}}><MessageIcon /></Avatar>
    }
    return (
      <Card style={{position: 'relative'}}>
        <div style={{position: 'absolute', right: 0}} onClick={() => {this.setState({isOpen: !this.state.isOpen})}}>
          {isOpen && <Button><KeyboardArrowUpIcon/></Button>}
          {!isOpen && <Button><KeyboardArrowDownIcon/></Button>}
        </div>
        <CardHeader
          title={comment}
          avatar={cardAvatar}
          subheader={subtitle}
          />
        <CardContent>
          {isOpen && <List>
            {alertData && <ListItem>
            <ListItemText
                primary={alertData.message}
                secondary="Message"/>
            </ListItem>
            }
            {alertData && <ListItem>
            <ListItemText
                primary={JSON.stringify(alertData.extra)}
                secondary="Alert Extra"/>
            </ListItem>
            }

            {task && <ListItem>
            <ListItemText
                primary={task.test_name}
                secondary="Test name"/>
            </ListItem>
            }
            {task && <ListItem>
            <ListItemText
                primary={JSON.stringify(task.arguments)}
                secondary="Test arguments"/>
            </ListItem>
            }

            <ListItem>
            <ListItemText
                primary={schedule}
                secondary="Schedule"/>
            </ListItem>

            <ListItem>
            <ListItemText
                primary={''+delay}
                secondary="Delay"/>
            </ListItem>

            <ListItem>
            <ListItemText
                primary={creationTime}
                secondary="Creation time"/>
            </ListItem>

            <ListItem>
            <ListItemText
                primary={targetCountries}
                secondary="Target countries"/>
            </ListItem>
            <ListItem>
            <ListItemText
                primary={targetPlatforms}
                secondary="Target platforms"/>
            </ListItem>
          </List>}
        </CardContent>
        <CardActions>
          {state !== 'deleted' && <Button onClick={() => {onDelete(id)}}><DeleteIcon/></Button>}
           <Button onClick={() => {alert('I do nothing')}}>Edit</Button>
        </CardActions>
      </Card>
    )
  }
}

class AdminJobsIndex extends React.Component {
  constructor (props) {
    console.log("Calling constructor")
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
    let req = this.state.session.createRequest({baseURL: process.env.ORCHESTRATE_URL})
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
    let req = this.state.session.createRequest({baseURL: process.env.ORCHESTRATE_URL})
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
          <Container style={{position: 'relative'}}>
            {jobList.map((job) => {
              return (
                <Flex>
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
                </Flex>
              )
            })}
            <div>
              <Flex column>
                {actionButtonOpen && <Box pt={2}>
                  <Link href='/admin/jobs/add_alert'><Button variant="fab" mini><MessageIcon/></Button></Link>
                </Box>}
                {actionButtonOpen && <Box pt={2}>
                  <Link href='/admin/jobs/add_task'><Button variant="fab" mini><AssignmentIcon/></Button></Link>
                </Box>}

                <Box pt={2} onClick={() => this.toggleAction()}>
                  {!actionButtonOpen && <Button variant="fab" color="primary" mini><AddIcon/></Button>}
                  {actionButtonOpen && <Button variant="fab" color="primary" mini><ClearIcon/></Button>}
                </Box>
              </Flex>
            </div>
          </Container>
        </div>
      </Layout>
    )
  }
}

export default AdminJobsIndex
