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
import Select from 'material-ui/Select'
import { MenuItem } from 'material-ui/Menu'

import MdDelete from 'react-icons/lib/md/delete'

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
  Text,
  InputWithIconButton
} from 'ooni-components'

import styled from 'styled-components'

import moment from 'moment'

const AddURLButton = styled(Button)`
  color: ${props => props.theme.colors.gray5};
  border-radius: 0;
  padding: 0;
  background-color: transparent;
  border-bottom: 1px solid ${props => props.theme.colors.gray1};
  text-align: left;
  text-transform: none;
  &:hover {
    background-color: transparent;
  color: ${props => props.theme.colors.gray6};
    border-bottom: 1px solid ${props => props.theme.colors.gray3};
  }
  &:active {
    background-color: transparent;
  color: ${props => props.theme.colors.gray7};
    border-bottom: 2px solid ${props => props.theme.colors.gray4};
  }
`

function getSteps() {
  return ['Choose experiment & targets', 'Repeat?', 'Sign & submit!'];
}

// This component is stolen from ooni/run.
// XXX we should probably move this into ooni/design
class AddURLsSection extends React.Component {

  constructor(props) {
    super(props)
    this.state = {
      urls: props.urls
    }
    this.handleDeleteURL = this.handleDeleteURL.bind(this)
    this.handleEditURL = this.handleEditURL.bind(this)
    this.handleKeyPress = this.handleKeyPress.bind(this)
    this.addURL = this.addURL.bind(this)
    this.urlRefs = new Map()
  }

  addURL() {
    let state = Object.assign({}, this.state)
    const idx = this.state.urls.length
    state.urls.push({value: 'http://', error: null, ref: null})
    this.props.onUpdatedURLs(state.urls)
    this.setState(state, () => {
      // This is a ghetto hax, that is a workaround for:
      // https://github.com/jxnblk/rebass/issues/329
      const urlInputs = document.getElementsByClassName('url-input')
      const target = urlInputs[urlInputs.length - 1]
      target.focus()
      target.setSelectionRange(7,7)
    })
  }

  handleKeyPress (e) {
    if (e.key === 'Enter') {
      this.addURL()
    }
  }

  handleDeleteURL(idx) {
    return ((event) => {
      let state = Object.assign({}, this.state)
      state.urls = state.urls
                        .filter((url, jdx) => jdx !== idx)
                        .map(url => Object.assign({}, url))
      this.setState(state)
      this.props.onUpdatedURLs(state.urls)
    }).bind(this)
  }

  handleEditURL(idx) {
    return ((event) => {
      const value = event.target.value
      let state = Object.assign({}, this.state)
      state.urls = state.urls.map(url => Object.assign({}, url))
      state.error = false
      let update = value.split(' ').map((line) => {
        let itm = {'value': line, 'error': null}
        if (!line.startsWith('https://') && !line.startsWith('http://')) {
          itm['error'] = 'URL must start with http:// or https://'
          state.error = true
        }
        return itm
      })
      state.urls.splice.apply(state.urls, [idx, 1].concat(update))
      this.setState(state)
    })
  }

  render() {
    const { onUpdatedURLs } = this.props
    const { urls } = this.state

    return (
      <Box w={2/3}>
      <Heading h={4} pb={1}>URLs</Heading>
        {urls.length == 0
        && <div>
          Click "Add URL" below to add a URL to test
          </div>
        }
        {urls.map((url, idx) => <div key={`url-${idx}`}>
          <InputWithIconButton
                className='url-input'
                value={url.value}
                icon={<MdDelete />}
                error={url.error}
                onKeyPress={this.handleKeyPress}
                onBlur={() => onUpdatedURLs(urls)}
                onChange={this.handleEditURL(idx)}
                onAction={this.handleDeleteURL(idx)} />
          </div>)}
        <div>
          <AddURLButton onClick={this.addURL}>
          + Add URL
          </AddURLButton>
        </div>
        </Box>
      )
    }
}

const CodeFormat = styled.div`
  font-family: monospace;
  word-wrap: break-word;
`

const ExperimentSign = ({data, signedExperiment, onSignedChange}) => {
  return (
    <div>
    <p>Copy paste the following into your terminal</p>
    <Heading h={4}>Sign</Heading>
    <CodeFormat>echo "{data}" | orchestrate sign</CodeFormat>
    <Heading h={4}>Paste</Heading>
    <TextField
      label="Signed Experiment"
      multiline
      fullWidth
      rowsMax="15"
      value={signedExperiment}
      onChange={onSignedChange}
    />
    </div>
  )
}

function b64EncodeUnicode(str) {
    // first we use encodeURIComponent to get percent-encoded UTF-8,
    // then we convert the percent encodings into raw bytes which
    // can be fed into btoa.
    return btoa(encodeURIComponent(str).replace(/%([0-9A-F]{2})/g,
        function toSolidBytes(match, p1) {
            return String.fromCharCode('0x' + p1);
    }));
}

class AdminNewExperiment extends React.Component {

  constructor (props) {
    super(props)
    this.state = {
      startMoment: moment(),
      repeatCount: 1,
      nettest: '',
      urls: [
        {value: 'http://', error: null}
      ],
      targetCountries: [],
      targetPlatforms: [],
      duration: {W: 1},
      inputSelectorOpen: false,
      signedExperiment: '',
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

    this.onNettestChange = this.onNettestChange.bind(this)
    this.onUrlsChange = this.onUrlsChange.bind(this)
    this.makeStringToSign = this.makeStringToSign.bind(this)
    this.onSignedChange = this.onSignedChange.bind(this)

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

    props.nettests = [
      {value: 'web_connectivity', label: 'Web Connectivity'},
      {value: 'facebook_messenger', label: 'Facebook Messenger'},
      {value: 'whatsapp', label: 'WhatsApp'},
      {value: 'telegram', label: 'Telegram'}
    ]

    return props
  }

  componentDidMount() {
    this.state.session.redirectIfInvalid()
  }

  onNettestChange ({target}) {
    this.setState({ nettest: target.value})
  }

  onUrlsChange (value) {
    this.setState({urls: value})
  }

  onSignedChange ({target}) {
    console.log(target)
    this.setState({ signedExperiment: target.value})
  }

  onDurationChange ({target}) {
    this.setState({ duration: target.value });
  }

  onRepeatChange (repeatCount) {
    this.setState({ repeatCount });
  }

  makeStringToSign() {
    const testName = this.state.nettest
    let probeCC = this.state.targetCountries.slice()
    if (probeCC.indexOf('any') !== -1) {
      probeCC = []
    }

    let args = {}
    if (testName === 'web_connectivity') {
      args.urls = this.state.urls.map(({value}) => ({url: value, code: 'XXX'})) // XXX resolve the category code
    }

    let jsonData = {
      'exp': moment().add(7, 'days').unix(), // XXX this should be set in relation to the schedule
      'iss': 'testing', // XXX change this for production usage
      'probe_cc': probeCC,
      'test_name': testName,
      'schedule': ToScheduleString({
                      duration: this.state.duration,
                      startMoment: this.state.startMoment,
                      repeatCount: this.state.repeatCount
                  }),
      'args': args
    }
    return b64EncodeUnicode(JSON.stringify(jsonData))
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
    let platforms = this.state.targetPlatforms.slice()
    if (platforms.indexOf('any') !== -1) {
      platforms = []
    }
    let countries = this.state.targetCountries.slice()
    if (countries.indexOf('any') !== -1) {
      countries = []
    }
    req.post('/api/v1/admin/experiment', {
      'schedule': ToScheduleString({
                      duration: this.state.duration,
                      startMoment: this.state.startMoment,
                      repeatCount: this.state.repeatCount
                  }),
      // XXX we currently don't set this
      'delay': 0,
      'signed_experiment': this.state.signedExperiment,
      'target': {
        'countries': countries,
        'platforms': platforms
      }
    }).then((res) => {
      this.setState({
       error: null,
       submitting: false
      })
      Router.push('/admin/experiments')
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
            <Heading h={2}>New Experiment</Heading>
            <Flex wrap>
              <Box w={1/3}>
              <Select
                value={this.state.nettest}
                onChange={this.onNettestChange}>
                {this.props.nettests.map(nt =>
                  <MenuItem value={nt.value} key={nt.value}>{nt.label}</MenuItem>)}
              </Select>
              </Box>

              {this.state.nettest === 'web_connectivity'
              && <AddURLsSection
                    urls={this.state.urls}
                    onUpdatedURLs={this.onUrlsChange} />}
              <Box w={1} pt={1}>
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
          <ExperimentSign
            data={this.makeStringToSign()}
            signedExperiment={this.state.signedExperiment}
            onSignedChange={this.onSignedChange} />
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

export default withStyles(styles)(AdminNewExperiment)
