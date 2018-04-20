import React from 'react'
import PropTypes from 'prop-types'
import Head from 'next/head'
import Router from 'next/router'

import { withStyles } from 'material-ui/styles'
import Button from 'material-ui/Button'
import Input, { InputLabel } from 'material-ui/Input'
import Checkbox from 'material-ui/Checkbox'
import { FormGroup, FormControlLabel } from 'material-ui/Form'
import Card, { CardHeader, CardContent, CardActions } from 'material-ui/Card'
import TextField from 'material-ui/TextField'
import List, { ListItem, ListItemText } from 'material-ui/List'

import Layout from '../../../components/layout'
import Session from '../../../components/session'
import {
  ToScheduleString
} from '../../../components/ui/schedule'

import TargetConfig from '../../../components/ui/jobs/TargetConfig'
import ScheduleConfig from '../../../components/ui/jobs/ScheduleConfig'

import {
  Flex, Box, Grid,
  Container,
  Heading
} from 'ooni-components'

import moment from 'moment'

class JobCreateConfirm extends React.Component {
  static propTypes = {
    startMoment: PropTypes.object,
    duration: PropTypes.object,
    repeatCount: PropTypes.number,
    targetCountries: PropTypes.array,
    targetPlatforms: PropTypes.array,
    urls: PropTypes.string,
    alertMessage: PropTypes.string,
    href: PropTypes.string,
    altHref: PropTypes.string
  }

  constructor(props) {
    super(props)
  }

  render() {
    const {
      startMoment,
      duration,
      repeatCount,
      alertMessage,
      href,
      altHref,
      targetCountries,
      targetPlatforms
    } = this.props

    const DurationCaption = ({duration, repeatCount, startMoment}) => (<div>
      <RepeatString duration={duration} repeatCount={repeatCount} />
      {ToScheduleString({
        duration: duration,
        startMoment: startMoment,
        repeatCount: repeatCount
      })}
    </div>
    )

    let startTimeCaption = startMoment.calendar()
    startTimeCaption += ' ('
    startTimeCaption += startMoment.toString()
    startTimeCaption += ')'

    return (
      <div>
        <CardHeader title="New Alert Summary" />
        <CardContent>
          <List>
          <ListItem>
            <ListItemText
            primary={alertMessage}
            secondary="Message" />
          </ListItem>
          <ListItem>
            <ListItemText
            primary={href}
            secondary="Link" />
          </ListItem>
          <ListItem>
            <ListItemText
            primary={startTimeCaption}
            secondary="Start time" />
          </ListItem>
          <ListItem>
            <ListItemText
            primary={<DurationCaption
                        duration={duration}
                        repeatCount={repeatCount}
                        startMoment={startMoment} />}
            secondary="Duration" />

          </ListItem>
          <ListItem>
            <ListItemText
            primary={targetCountries.join(',')}
            secondary="Target countries" />
          </ListItem>
          <ListItem>
            <ListItemText
            primary={targetPlatforms.join(',')}
            secondary="Target platforms" />
          </ListItem>
          </List>
        </CardContent>
      </div>
    )
  }
}

class AdminJobsAdd extends React.Component {

  constructor (props) {
    super(props)
    this.state = {
      startMoment: moment(),
      repeatCount: 1,
      alertMessage: '',
      href: '',
      altHref: '',
      targetCountries: [],
      targetPlatforms: [],
      duration: {W: 1},
      inputSelectorOpen: false,
      submitted: false,
      comment: '',
      session: new Session(),
      finalized: null
    }

    this.onTargetCountryChange = this.onTargetCountryChange.bind(this)
    this.onTargetPlatformChange = this.onTargetPlatformChange.bind(this)
    this.onDurationChange = this.onDurationChange.bind(this)
    this.onRepeatChange = this.onRepeatChange.bind(this)
    this.onSubmit = this.onSubmit.bind(this)
    this.onEdit = this.onEdit.bind(this)
    this.onAdd = this.onAdd.bind(this)
    this.onMessageChange = this.onMessageChange.bind(this)
    this.onHrefChange = this.onHrefChange.bind(this)
    this.onAltHrefChange = this.onAltHrefChange.bind(this)
  }

  static async getInitialProps ({req, res}) {
    // XXX get these from an API call
    const cat_codes = require('../../../static/category-codes.json')
    const countries_alpha2 = require('../../../static/countries-alpha2.json')

    let props = {}
    props.countries = [
      {value: "any", label: 'All'}
    ]
    for (let alpha2 in countries_alpha2) {
      props.countries.push({
        value: alpha2,
        label: countries_alpha2[alpha2]
      })
    }

    props.platforms = [
      {value: 'any', label: 'All'},
      {value: 'android', label: 'Android'},
      {value: 'ios', label: 'iOS'},
      {value: 'linux', label: 'Linux'},
      {value: 'macos', label: 'macOS'},
      {value: 'lepidopter', label: 'Lepidopter'},
    ]

    return props
  }

  componentDidMount() {
    this.state.session.redirectIfInvalid()
  }

  onMessageChange ({target}) {
    this.setState({ alertMessage: target.value})
  }
  onHrefChange ({target}) {
    this.setState({ href: target.value})
  }
  onAltHrefChange ({target}) {
    this.setState({ altHref: target.value})
  }

  onDurationChange ({target}) {
    this.setState({ duration: target.value });
  }

  onRepeatChange (repeatCount) {
    this.setState({ repeatCount });
  }

  onTargetCountryChange (valueList) {
    let value = valueList.map(x => x.value)
    if (value.indexOf('any') != -1) {
      if (this.state.targetCountries.indexOf('any') != -1) {
        // If any was already there we remove it
        value.pop('any')
      } else {
        // Otherwise we clear everything else
        value = ['any']
      }
    }
    this.setState({ targetCountries: value })
  }

  onTargetPlatformChange (valueList) {
    let value = valueList.map(x => x.value)
    if (value.indexOf('any') != -1) {
      if (this.state.targetPlatforms.indexOf('any') != -1) {
        // If any was already there we remove it
        value.pop('any')
      } else {
        // Otherwise we clear everything else
        value = ['any']
      }
    }
    this.setState({ targetPlatforms: value })
  }

  onSubmit () {
    console.log('on submit')
    this.setState({
      submitted: true
    })
  }

  onEdit () {
    this.setState({
      submitted: false
    })
  }

  onAdd () {
    console.log('on add')
    let req = this.state.session.createRequest({baseURL: process.env.ORCHESTRATE_URL})
    let alertExtra = {}
    if (this.state.href != '') {
      alertExtra['href'] = this.state.href
      alertExtra['alt_hrefs'] = []
      if (this.state.altHref != '') {
        this.state.altHref.split("\n").forEach((v) => {
          if (v != '') {
            alertExtra['alt_hrefs'].push(v)
          }
        })
      }
    }
    let platforms = this.state.targetPlatforms.slice()
    if (platforms.indexOf('any') !== -1) {
      platforms = []
    }
    let countries = this.state.targetCountries.slice()
    if (countries.indexOf('any') !== -1) {
      countries = []
    }
    req.post('/api/v1/admin/job', {
      'schedule': ToScheduleString({
                      duration: this.state.duration,
                      startMoment: this.state.startMoment,
                      repeatCount: this.state.repeatCount
                  }),
      // XXX we currently don't set this
      'delay': 0,
      'comment': this.state.alertMessage,
      'alert': {
        'message': this.state.alertMessage,
        'extra': alertExtra,
      },
      'target': {
        'countries': countries,
        'platforms': platforms
      }
    }).then((res) => {
      this.setState({
        finalized: {
          error: null
        },
        submitted: true
      })
      Router.push('/admin/jobs')
    }).catch((err) => {
      this.setState({
        finalized: {
          error: err
        },
        submitted: true
      })
    })
  }

  render () {
    const {
      submitted,
      startMoment,
      repeatCount,
      alertMessage,
      targetCountries,
      targetPlatforms,
      duration,
      finalized
    } = this.state


    return (
      <Layout title="Add Jobs">
        <Head>
          <title>Jobs - OONI Proteus</title>
          <link href="/static/vendor/react-select.css" rel="stylesheet" />
        </Head>

          <div>
          <Container>
            {submitted &&
              <Card>
              {finalized && finalized.error === null &&
                <CardHeader title="Job created successfully!" />
              }
              {finalized && finalized.error !== null &&
                <div>
                <CardHeader title="Job creation error" />
                <p>{finalized.error.toString()}</p>
                <CardActions>
                <Button
                  raised
                  onClick={this.onEdit}>Edit</Button>
                <Button
                  raised
                  onClick={this.onAdd}>Retry</Button>
                </CardActions>
                </div>
              }
              {!finalized && <div>
                <JobCreateConfirm
                  startMoment={this.state.startMoment}
                  duration={this.state.duration}
                  repeatCount={this.state.repeatCount}
                  duration={this.state.duration}

                  alertMessage={this.state.alertMessage}
                  href={this.state.href}
                  altHref={this.state.altHref}

                  targetCountries={this.state.targetCountries}
                  targetPlatforms={this.state.targetPlatforms}
                  urls={this.state.urls}
                  comment={this.state.comment}
                />
                <CardActions>
                <Button
                  onClick={this.onEdit}>Edit</Button>
                <Button
                  onClick={this.onAdd}>Add</Button>
                </CardActions>
              </div>}

              </Card>
            }
          </Container>
          {!submitted &&
          <Container>
            <Card>
              <CardHeader  title="New Alert" />
              <CardContent>
              <Flex wrap>
                <Box w={1}>
                <InputLabel>Message</InputLabel>
                <Input
                  fullWidth
                  onChange={this.onMessageChange}
                  placeholder="make it short"
                  type="text" />
                </Box>
                <Box w={1} pt={3}>
                <InputLabel>Link</InputLabel>
                <Input
                  fullWidth
                  onChange={this.onHrefChange}
                  placeholder="https://msg.ooni.io/xxx"
                  type="text" />
                </Box>
                <Box w={1} pt={1}>
                <Input
                  fullWidth
                  onChange={this.onAltHrefChange}
                  placeholder="https://cloudfront.com/foo/bar/xxx"
                  multiline
                  type="text" />
                </Box>
              </Flex>
              <hr/>

              <Heading h={2}>Target</Heading>
              <TargetConfig
                countries={this.props.countries}
                targetCountries={this.state.targetCountries}
                onTargetCountryChange={this.onTargetCountryChange}
                platforms={this.props.platforms}
                targetPlatforms={this.state.targetPlatforms}
                onTargetPlatformChange={this.onTargetPlatformChange}
              />

              <hr />

              <Heading h={2}>Schedule</Heading>
              <ScheduleConfig
                startMoment={this.state.startMoment}
                onStartMomentChange={(startMoment) => {
                  this.setState({ startMoment })
                }}
                repeatCount={this.state.repeatCount}
                onRepeatChange={this.onRepeatChange}
                duration={this.state.duration}
                onDurationChange={this.onDurationChange} />

              </CardContent>
              <CardActions>
                <Button
                  raised
                  onClick={this.onSubmit}
                  style={{marginLeft: 20}}>Add</Button>
              </CardActions>
            </Card>
            </Container>}
        </div>
      </Layout>
    )
  }
}

const styles = theme => ({
  root: {
    display: 'flex',
  },
  formControl: {
    margin: theme.spacing.unit * 3,
  },
  group: {
    margin: `${theme.spacing.unit}px 0`,
  },
});

export default withStyles(styles)(AdminJobsAdd)
