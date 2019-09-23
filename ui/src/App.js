import React from 'react';
import './App.css';
import 'bootstrap/dist/css/bootstrap.css';


class AppCard extends React.Component {
    componentDidMount() {
        this.callApi()
            .then(res => this.setState({ response: res.express }))
            .catch(err => console.log(err));
    }

    callApi = async () => {
        const response = await fetch('/api/hello');
        const body = await response.json();
        if (response.status !== 200) throw Error(body.message);

        return body;
    };

    render() {
        return (
            <div className="col-md-2">
                <div className="card text-white mb-3" className={this.props.fixed ? 'bg-success' : 'bg-danger'}>
                    <div className="card-body">
                        <h5 className="card-title">{this.props.name}</h5>
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


function App() {
    return (
        <div className="container-fluid">
            <AppList apps={[{name: "app2", fixed: false}, {name: "app 1", fixed: true}]}/>
        </div>
    );
}

export default App;
