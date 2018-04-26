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
import Stepper, { Step, StepLabel } from 'material-ui/Stepper'
import { CircularProgress } from 'material-ui/Progress'

import Layout from '../../../components/layout'
import Session from '../../../components/session'
import {
    RepeatString,
    ToScheduleString
} from '../../../components/ui/schedule'

import TargetConfig from '../../../components/ui/jobs/TargetConfig'
import ScheduleConfig from '../../../components/ui/jobs/ScheduleConfig'

import {
  Flex, Box, Grid,
  Container,
  Heading,
  Text
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
        <Heading h={2}>New Alert Review</Heading>

        <Flex wrap>
        <Box w={1} pb={3}>
        <InputLabel>Message</InputLabel>
        <Text>{alertMessage}</Text>
        </Box>

        <Box w={1} pb={3}>
        <InputLabel>Link</InputLabel>
        <Text>{href}</Text>
        </Box>

        <Box w={1} pb={3}>
        <InputLabel>Target countries</InputLabel>
        <Text>{targetCountries.join(',')}</Text>
        </Box>
        <Box w={1} pb={3}>
        <InputLabel>Target platforms</InputLabel>
        <Text>{targetPlatforms.join(',')}</Text>
        </Box>

        <Box w={1} pb={3}>
        <InputLabel>Start Time</InputLabel>
        <Text>{startTimeCaption}</Text>
        </Box>

        <Box w={1} pb={3}>
        <InputLabel>Duration</InputLabel>
        <DurationCaption
          duration={duration}
          repeatCount={repeatCount}
          startMoment={startMoment} />
        </Box>

        </Flex>
      </div>
    )
  }
}

function getSteps() {
  return ['Choose message & targets', 'Repeat?', 'Review & submit!'];
}

const getStepContent = (idx) => {
  switch (idx) {
    case 0:
      return 'Choose message & targets';
    case 1:
      return 'Repeat?';
    case 2:
      return 'Review & submit!';
    default:
      return 'Unknown step';
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
      comment: '',
      session: new Session(),
      activeStep: 0,
      skipped: new Set(),
      error: null,
      submitting: false
    }

    this.onTargetCountryChange = this.onTargetCountryChange.bind(this)
    this.onTargetPlatformChange = this.onTargetPlatformChange.bind(this)
    this.onDurationChange = this.onDurationChange.bind(this)
    this.onRepeatChange = this.onRepeatChange.bind(this)

    this.onMessageChange = this.onMessageChange.bind(this)
    this.onHrefChange = this.onHrefChange.bind(this)
    this.onAltHrefChange = this.onAltHrefChange.bind(this)

    this.onSubmit = this.onSubmit.bind(this)
    this.isStepSkipped = this.isStepSkipped.bind(this)
    this.handleNext = this.handleNext.bind(this)
    this.handleBack = this.handleBack.bind(this)
    this.handleSkip = this.handleSkip.bind(this)
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
    this.setState({
      submitting: true
    })

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
    req.post('/api/v1/admin/alert', {
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
       error: null,
       submitting: false
      })
      Router.push('/admin/alerts')
    }).catch((error) => {
      this.setState({
        submitting: false,
        error
      })
    })
  }

  isStepOptional (step) {
    return step === 1;
  }

  isStepSkipped(step) {
    return this.state.skipped.has(step)
  }

  handleNext() {
    const { activeStep } = this.state;
    let { skipped } = this.state;
    if (this.isStepSkipped(activeStep)) {
      skipped = new Set(skipped.values());
      skipped.delete(activeStep);
    }
    this.setState({
      activeStep: activeStep + 1,
      error: null,
      skipped,
    })
  }

  handleBack() {
    const { activeStep } = this.state;
    this.setState({
      activeStep: activeStep - 1,
    })
  }

  handleSkip() {
    const { activeStep } = this.state;
    if (!this.isStepOptional(activeStep)) {
      throw new Error("You can't skip a step that isn't optional.");
    }
    const skipped = new Set(this.state.skipped.values());
    skipped.add(activeStep)
    this.setState({
      activeStep: this.state.activeStep + 1,
      skipped,
    });
  }

  render () {
    const {
      submitting,
      startMoment,
      repeatCount,
      alertMessage,
      targetCountries,
      targetPlatforms,
      duration,
      activeStep,
      error
    } = this.state

    const steps = getSteps()

    return (
      <Layout title="Add Jobs">
        <Head>
          <title>Jobs - OONI Orchestra</title>
          <link href="/static/vendor/react-select.css" rel="stylesheet" />
        </Head>

          <div>
          <div>
          <Stepper activeStep={activeStep}>
            {steps.map((label, index) => {
              const props = {}
              if (this.isStepSkipped(index)) {
                props.completed = false
              }
              return (
                <Step key={label} {...props}>
                  <StepLabel>{label}</StepLabel>
                </Step>
              );
            })}
          </Stepper>
          </div>

          {error &&
            <div>
            <p>Job creation error</p>
            <p>{error.toString()}</p>
            </div>
          }

          {submitting && <CircularProgress />}

          {activeStep === 0 && <Container>
            <Heading h={2}>New alert</Heading>
            <Heading color='red' h={4}>Important</Heading>
            <p>Before you begin you should go to <a
  href="https://msg.ooni.io/">https://msg.ooni.io/</a> and create a post that is
  to be linked via this alert.
            </p>
            <p style={{paddingBottom: '20px'}}>Return to this page once you have done that</p>
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
          </Container>}

          {activeStep === 1 && <Container>
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
          </Container>}

          {activeStep === 2 && <Container>
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
          </Container>}

          <Container pt={3}>
          <Button
            disabled={activeStep === 0}
            onClick={this.handleBack}
            >Back</Button>
          {this.isStepOptional(activeStep) && (
            <Button
              variant="raised"
              color="primary"
              onClick={this.handleSkip}
            >
              Skip
            </Button>
          )}
          {activeStep === steps.length - 1 ? <Button
              variant="raised"
              color="primary"
              onClick={this.onSubmit}
            >Submit</Button>
            : <Button
            variant="raised"
            disabled={submitting === true}
            color="primary"
            onClick={this.handleNext}>Next</Button>}
          </Container>

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
})

export default withStyles(styles)(AdminJobsAdd)
