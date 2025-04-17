import React, { useState, useEffect, useRef } from "react";

const DropDownMenu = ({ options, dropdownId, onSelect }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [selected, setSelected] = useState("Select an option");
  const dropdownRef = useRef(null);

  // Handle clicking outside to close dropdown
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  const handleSelect = (option, value) => {
    setSelected(option);
    setIsOpen(false);
    if (onSelect) {
      onSelect(value !== undefined ? value : option);
    }
  };

  return (
    <div className="relative inline-block text-left w-full" id={`dropdown-${dropdownId}`} ref={dropdownRef}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="inline-flex justify-between w-full px-4 py-2 bg-white border border-gray-300 rounded-md shadow-sm hover:bg-gray-50 focus:outline-none"
      >
        <span className="truncate">{selected}</span>
        <svg
          className={`w-5 h-5 ml-2 transition-transform duration-200 ${isOpen ? 'transform rotate-180' : ''}`}
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 20 20"
          fill="currentColor"
        >
          <path
            fillRule="evenodd"
            d="M5.23 7.21a.75.75 0 011.06.02L10 11.584l3.71-4.354a.75.75 0 111.14.976l-4.25 5a.75.75 0 01-1.14 0l-4.25-5a.75.75 0 01.02-1.06z"
            clipRule="evenodd"
          />
        </svg>
      </button>

      {isOpen && (
        <div className="absolute z-50 w-full mt-1 origin-top-right bg-white border border-gray-200 divide-y divide-gray-100 rounded-md shadow-lg max-h-60 overflow-y-auto">
          {options.map((option, index) => {
            // Handle both simple strings and objects with label/value
            const displayText = typeof option === 'object' ? option.label || option.name : option;
            const value = typeof option === 'object' ? option.value || option.id : option;
            
            return (
              <div
                key={index}
                onClick={() => handleSelect(displayText, value)}
                className="px-4 py-2 cursor-pointer hover:bg-gray-100 text-sm"
              >
                {displayText}
              </div>
            );
          })}
        </div>
      )}
      
      {/* Hidden input to store the selected value */}
      <input type="hidden" id={dropdownId} value={typeof selected === 'object' ? selected.value || selected.id : selected} />
    </div>
  );
};

export default DropDownMenu;