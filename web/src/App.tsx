import { useState } from "react";
import { Client } from "../api";
import * as api_system_city from "../api/api_system_city";
import "./App.css";
import reactLogo from "./assets/react.svg";
import viteLogo from "/vite.svg";

const client = new Client("http://localhost:8080");

function App() {
  const [cityList, setCityList] = useState<api_system_city.CityList>();

  return (
    <>
      <div>
        <a href="https://vite.dev" target="_blank">
          <img src={viteLogo} className="logo" alt="Vite logo" />
        </a>
        <a href="https://react.dev" target="_blank">
          <img src={reactLogo} className="logo react" alt="React logo" />
        </a>
      </div>
      <h1>Vite + React</h1>
      <div className="card">
        <button
          onClick={async () => {
            const ret = await client.System.City.GetCityList("cn");
            console.log(ret);
            setCityList(ret);
          }}
        >
          {JSON.stringify(cityList)}
        </button>
        <p>
          Edit <code>src/App.tsx</code> and save to test HMR
        </p>
      </div>
      <p className="read-the-docs">
        Click on the Vite and React logos to learn more
      </p>
    </>
  );
}

export default App;
