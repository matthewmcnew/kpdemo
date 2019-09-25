import React from 'react';
import './App.css';
import 'bootstrap/dist/css/bootstrap.css';
import {fetchImages} from "./images";


class AppCard extends React.Component {
    buildMetadata;

    render() {
        return (
            <div className="col-sm-3 py-2">
                <div className={`card h-100 ${this.color()}`}>
                    <div className="card-body">
                        <h4 className="card-title">{this.props.name}</h4>
                        <p className="card-text">Namespace:{this.props.namespace}</p>
                        <p className="card-text">
                            {
                                this.description()
                            }
                        </p>
                    </div>
                </div>
            </div>);
    }

    color() {
        if (this.props.status === "True") {
            return "bg-success"
        } else if (this.props.status === "False") {
            return "bg-danger"
        }
        return "bg-secondary"
    }

    description() {
        if (this.props.status === "True") {
            return this.buildpacks().map((item) =>
                <small>{item.key}:{item.version}<br></br></small>
            )
        }

        return this.props.remaining + "/9"

    }

    buildpacks() {
        if (this.props.buildMetadata == null) {
            return []
        }
        return this.props.buildMetadata
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
        setInterval(() => fetchImages()
            .then(res => this.setState({images: res})).catch(err => console.log(err)), 1000)
    }

    render() {
        console.log(this.state)

        return (
            <div className="container-fluid">
                <br/>
                <AppList apps={this.state.images}/>
            </div>
        );
    }
}

export default App;
