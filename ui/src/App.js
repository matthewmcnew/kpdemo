import React from 'react';
import './App.css';
import 'bootstrap/dist/css/bootstrap.css';


class AppCard extends React.Component {
    render() {
        return (
            <div className="col-md-2">
                <div className="card text-white mb-3" className={this.props.Ready ? 'bg-success' : 'bg-danger'}>
                    <div className="card-body">
                        <h5 className="card-title">{this.props.Name}</h5>
                        <p className="card-text">This will show metadata of pod</p>
                    </div>
                </div>
            </div>);
    }

}

class AppList extends React.Component {
    render() {
        return (
            <div className="row">
                {this.props.apps.map((app) => <AppCard {...app}/>)}
            </div>
        );
    }
}


class App extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            images: []
        }
    }

    componentDidMount() {
        setInterval(() => this.callApi()
            .then(res => this.setState({images: res})).catch(err => console.log(err)), 1000)
    }

    callApi = async () => {
        const response = await fetch('/images');
        debugger;
        const body = await response.json();
        if (response.status !== 200) throw Error(body.message);
        console.log(body)
        return body;
    };

    render() {
        console.log(this.state)

        return (
            <div className="container-fluid">
                <AppList apps={this.state.images}/>
            </div>
        );
    }
}

export default App;
