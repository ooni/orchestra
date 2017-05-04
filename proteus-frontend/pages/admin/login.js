import Page from '../../components/page'
import Session from '../../components/session'
import Layout from '../../components/layout'

import RaisedButton from 'material-ui/RaisedButton'
import TextField from 'material-ui/TextField'

import { Flex, Box, Grid } from 'reflexbox'

export default class AdminLogin extends Page {

  constructor(props) {
    super(props)
    this.state = {
      username: '',
      password: ''
    }
    this.onSubmit = this.onSubmit.bind(this)
    this.onUsernameChange = this.onUsernameChange.bind(this)
    this.onPasswordChange = this.onPasswordChange.bind(this)
  }

  onUsernameChange(username) {
    this.setState({username})
  }
  onPasswordChange(password) {
    this.setState({password})
  }

  async onSubmit(event) {
    event.preventDefault()
    const session = new Session()
    session.login(this.state.username, this.state.password)
      .then(() => {
        console.log('logged in...')
      })
      .catch(err => {
        console.log('failed to login...')
        console.log(err)
      })
  }

  render() {
    return (
      <Layout>
        <div>
          <Grid col={3} px={2}>
            <TextField
              hintText={`your username`}
              floatingLabelText='Username'
              value={this.state.username}
              onChange={(event, value) => this.onUsernameChange(value)}
            />
          </Grid>
          <br/>
          <Grid col={3} px={2}>
            <TextField
              hintText={`your username`}
              floatingLabelText='Password'
              type="password"
              value={this.state.password}
              onChange={(event, value) => this.onPasswordChange(value)}
            />
          </Grid>
          <br/>
          <RaisedButton
            onTouchTap={this.onSubmit}
            label='Login' style={{marginLeft: 20}}/>
        </div>
      </Layout>
    )
  }
}
