import Router from 'next/router'

import Session from '../../components/session'
import Layout from '../../components/layout'

import Button from 'react-toolbox/lib/button/Button'
import Input from 'react-toolbox/lib/input/Input'
import Card from 'react-toolbox/lib/card/Card'
import CardActions from 'react-toolbox/lib/card/CardActions'
import CardTitle from 'react-toolbox/lib/card/CardTitle'

import { Flex, Box, Grid } from 'reflexbox'

export default class AdminLogin extends React.Component {

  constructor(props) {
    super(props)
    this.state = {
      username: '',
      password: '',
      error: null
    }
    this.onSubmit = this.onSubmit.bind(this)
    this.onUsernameChange = this.onUsernameChange.bind(this)
    this.onPasswordChange = this.onPasswordChange.bind(this)
  }

  onUsernameChange(username) {
    console.log("Setting username to", username)
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
        let redirectPath = '/'
        this.setState({
          error: null
        })
        if (Router.query.from) {
          redirectPath = Router.query['from']
        }
        Router.push(redirectPath)
      })
      .catch(err => {
        this.setState({
          error: {
            'message': 'Failed to login',
            'debug_error': err
          }
        })
      })
  }

  render() {
    const {
      error
    } = this.state

    return (
      <Layout>
        <Card style={{width: "600px", margin: "auto", padding: '20px'}}>
          {error !== null && <p>{error.message}</p>}
          <CardTitle title="Authorization required" style={{color: "black"}} />
          <Input
            type='text'
            hint='your username'
            label='Username'
            value={this.state.username}
            onChange={(value) => this.onUsernameChange(value)} />
          <Input
            type='password'
            hint='your password'
            label='Password'
            value={this.state.password}
            onChange={(value) => this.onPasswordChange(value)}/>
          <CardActions>
          <Button
            onClick={this.onSubmit}
            label='Login' />
          </CardActions>
        </Card>
      </Layout>
    )
  }
}
