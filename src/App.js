import React from "react";
import {BrowserRouter as Router, Routes,Route} from "react-router-dom";
import Groups from "./components/Groups";
import Group from "./components/Group";

const App = () =>{
  return (
    <>
      <Router>
        <Routes>
          <Route path="/" element={<Groups/>}/>
          <Route path="/group/:groupId/:groupName" element={<Group/>}/>
        </Routes>
      </Router>
    </>
  )
}

export default App;